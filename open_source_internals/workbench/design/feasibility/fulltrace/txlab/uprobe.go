package main

import (
	"bufio"
	"debug/dwarf"
	"debug/elf"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// 이 파일은 관측 정의(probes.txt)를 커널의 uprobe 정의 줄로 바꿉니다. 함수의 위치와 구조체 필드의 위치를 손으로 적지 않고
// 실행 파일의 DWARF에서 찾습니다. 같은 이름의 필드가 빌드마다 옮겨 가도 정의는 그대로 둘 수 있게 하려는 것입니다.
//
// 관측 정의 한 줄: "이름 함수[%return] 필드=식 ..."
//   식: 시작은 $arg1~$arg8(함수 인자), $retval(반환값), @전역변수. 그 뒤로 ->필드(포인터를 따라간 구조체의 필드),
//       .필드(구조체 안의 필드), [N](배열이나 포인터의 N번째)를 잇습니다. 예: $arg1->t_data->t_choice.t_heap.t_xmin
//   결과의 형식(u8·u16·u32·u64·s32·x64)은 마지막 값의 DWARF 형식에서 정합니다.
// 커널 문법(Documentation/trace/uprobetracer.rst): 인자는 %x0~%x7(arm64), 메모리 읽기는 +OFS(FETCHARG),
// 실행 파일 안 주소는 @+OFFSET입니다. 이 실행 파일은 코드 구간이 파일 위치 0에서 시작하므로 기호의 주소가 곧 파일 안 위치이고,
// @+OFFSET은 커널이 "걸린 자리 - 정의의 위치 + OFFSET"으로 계산하므로 전역 변수는 그 주소를 그대로 씁니다.

// probeDef 는 관측 정의 한 줄입니다.
type probeDef struct {
	Name, Func string
	Return     bool
	Fields     [][2]string
}

// dwarfIndex 는 실행 파일에서 찾은 함수와 전역 변수입니다.
type dwarfIndex struct {
	d       *dwarf.Data
	funcs   map[string]*dwarfFunc
	globals map[string]dwarfVar
}

type dwarfFunc struct {
	Low    uint64
	Params []dwarf.Type
	Ret    dwarf.Type
}

type dwarfVar struct {
	Addr uint64
	Type dwarf.Type
}

// loadDwarf 는 실행 파일의 DWARF를 한 번 훑어, 정의(DW_AT_low_pc가 있는 것)가 있는 함수와 주소가 정해진 전역 변수를 모읍니다.
// 같은 이름의 static 함수가 여러 파일에 있으면 처음 것을 쓰고 이름을 dup에 남깁니다.
//
// Errors: 파일을 열지 못하거나 DWARF가 없으면 오류를 돌려줍니다.
func loadDwarf(bin string, want map[string]bool) (*dwarfIndex, []string, error) {
	f, err := elf.Open(bin)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	d, err := f.DWARF()
	if err != nil {
		return nil, nil, err
	}
	idx := &dwarfIndex{d: d, funcs: map[string]*dwarfFunc{}, globals: map[string]dwarfVar{}}
	var dup []string
	r := d.Reader()
	var cur *dwarfFunc
	type declInfo struct {
		name   string
		typ    dwarf.Offset
		hasTyp bool
	}
	decls := map[dwarf.Offset]declInfo{}
	for {
		e, err := r.Next()
		if err != nil {
			return nil, nil, err
		}
		if e == nil {
			break
		}
		switch e.Tag {
		case dwarf.TagCompileUnit:
			cur = nil
		case dwarf.TagSubprogram:
			cur = nil
			name, _ := e.Val(dwarf.AttrName).(string)
			low, ok := e.Val(dwarf.AttrLowpc).(uint64)
			if !want[name] || !ok {
				if e.Children {
					r.SkipChildren()
				}
				continue
			}
			if _, seen := idx.funcs[name]; seen {
				dup = append(dup, name)
				r.SkipChildren()
				continue
			}
			fn := &dwarfFunc{Low: low}
			if off, ok := e.Val(dwarf.AttrType).(dwarf.Offset); ok {
				fn.Ret, _ = d.Type(off)
			}
			idx.funcs[name] = fn
			cur = fn
			if !e.Children {
				cur = nil
			}
		case dwarf.TagFormalParameter:
			if cur != nil {
				if off, ok := e.Val(dwarf.AttrType).(dwarf.Offset); ok {
					t, _ := d.Type(off)
					cur.Params = append(cur.Params, t)
				}
			}
		case dwarf.TagVariable:
			// 헤더에서 extern으로 선언한 전역은 정의 쪽에 이름이 없고 DW_AT_specification이 선언을 가리킵니다(MyProcPid가 그렇습니다).
			// 그래서 선언의 이름과 형식을 기억해 두었다가 정의에 씁니다.
			name, _ := e.Val(dwarf.AttrName).(string)
			typOff, hasTyp := e.Val(dwarf.AttrType).(dwarf.Offset)
			if spec, ok := e.Val(dwarf.AttrSpecification).(dwarf.Offset); ok {
				d := decls[spec]
				name = d.name
				if !hasTyp {
					typOff, hasTyp = d.typ, d.hasTyp
				}
			} else if decl, _ := e.Val(dwarf.AttrDeclaration).(bool); decl {
				decls[e.Offset] = declInfo{name, typOff, hasTyp}
			}
			if !want["@"+name] {
				continue
			}
			loc, ok := e.Val(dwarf.AttrLocation).([]byte)
			// DW_OP_addr(0x03) 뒤의 8바이트가 주소입니다. 다른 꼴의 위치(지역 변수 등)는 쓰지 않습니다.
			if !ok || len(loc) != 9 || loc[0] != 0x03 {
				continue
			}
			addr := uint64(0)
			for i := 8; i >= 1; i-- {
				addr = addr<<8 | uint64(loc[i])
			}
			var t dwarf.Type
			if hasTyp {
				t, _ = d.Type(typOff)
			}
			if _, seen := idx.globals[name]; !seen {
				idx.globals[name] = dwarfVar{Addr: addr, Type: t}
			}
		default:
			// 함수의 직접 자식만 인자입니다. 안쪽 블록(TagLexDwarfBlock 등)에 들어가면 그 뒤 인자는 없습니다.
			if e.Tag != 0 && cur != nil && e.Tag != dwarf.TagFormalParameter {
				cur = nil
			}
		}
	}
	return idx, dup, nil
}

// fetch 는 식을 따라가는 중간 결과입니다. lvalue가 거짓이면 expr이 값 자체(레지스터나 이미 읽은 메모리)이고,
// 참이면 값은 메모리에 있고 그 주소는 base의 값 + off입니다. base가 비어 있으면 실행 파일 안 고정 주소(off)입니다.
type fetch struct {
	lvalue bool
	expr   string
	base   string
	off    int64
	typ    dwarf.Type
}

// value 는 lvalue를 메모리 읽기 식으로 바꿉니다.
func (f fetch) value() string {
	if !f.lvalue {
		return f.expr
	}
	if f.base == "" {
		return fmt.Sprintf("@+0x%x", f.off)
	}
	return fmt.Sprintf("+%d(%s)", f.off, f.base)
}

// strip 은 typedef와 const·volatile을 벗깁니다.
func strip(t dwarf.Type) dwarf.Type {
	for {
		switch v := t.(type) {
		case *dwarf.TypedefType:
			t = v.Type
		case *dwarf.QualType:
			t = v.Type
		default:
			return t
		}
	}
}

var accessRe = regexp.MustCompile(`^(->|\.)(\w+)|^\[(\d+)\]`)

// resolve 는 식 하나를 커널의 인자 문법과 형식으로 바꿉니다.
//
// Errors: 인자 번호가 범위 밖이거나, 필드가 없거나, 포인터가 아닌 것을 따라가려 하면 오류를 돌려줍니다.
func (idx *dwarfIndex) resolve(fn *dwarfFunc, expr string) (string, error) {
	// ":string"로 끝나면 식의 값(char 포인터)이 가리키는 NUL 끝 문자열을 읽습니다. 예: sql=$arg1:string
	if base, ok := strings.CutSuffix(expr, ":string"); ok {
		v, err := idx.resolve(fn, base)
		if err != nil {
			return "", err
		}
		if !strings.HasSuffix(v, ":x64") {
			return "", fmt.Errorf("%s: 문자열은 포인터 값에서만 읽습니다", expr)
		}
		return "+0(" + strings.TrimSuffix(v, ":x64") + "):string", nil
	}
	var cur fetch
	rest := expr
	switch {
	case strings.HasPrefix(rest, "$retval"):
		cur = fetch{expr: "$retval", typ: fn.Ret}
		rest = rest[len("$retval"):]
	case strings.HasPrefix(rest, "$arg"):
		m := regexp.MustCompile(`^\$arg(\d)`).FindStringSubmatch(rest)
		if m == nil {
			return "", fmt.Errorf("%s: 인자 꼴이 아닙니다", expr)
		}
		n, _ := strconv.Atoi(m[1])
		if n < 1 || n > len(fn.Params) || n > 8 {
			return "", fmt.Errorf("%s: 이 함수의 인자는 %d개입니다", expr, len(fn.Params))
		}
		cur = fetch{expr: fmt.Sprintf("%%x%d", n-1), typ: fn.Params[n-1]}
		rest = rest[len(m[0]):]
	case strings.HasPrefix(rest, "@"):
		m := regexp.MustCompile(`^@(\w+)`).FindStringSubmatch(rest)
		g, ok := idx.globals[m[1]]
		if !ok {
			return "", fmt.Errorf("%s: 전역 변수 %s를 찾지 못했습니다", expr, m[1])
		}
		cur = fetch{lvalue: true, off: int64(g.Addr), typ: g.Type}
		rest = rest[len(m[0]):]
	default:
		return "", fmt.Errorf("%s: $argN, $retval, @전역변수로 시작해야 합니다", expr)
	}
	for rest != "" {
		m := accessRe.FindStringSubmatch(rest)
		if m == nil {
			return "", fmt.Errorf("%s: %q를 읽지 못했습니다", expr, rest)
		}
		rest = rest[len(m[0]):]
		t := strip(cur.typ)
		switch {
		case m[1] == "->":
			p, ok := t.(*dwarf.PtrType)
			if !ok {
				return "", fmt.Errorf("%s: ->%s 앞이 포인터가 아닙니다(%s)", expr, m[2], t)
			}
			st, ok := strip(p.Type).(*dwarf.StructType)
			if !ok {
				return "", fmt.Errorf("%s: ->%s 앞이 구조체 포인터가 아닙니다", expr, m[2])
			}
			fld := findField(st, m[2])
			if fld == nil {
				return "", fmt.Errorf("%s: %s에 %s가 없습니다", expr, st.StructName, m[2])
			}
			cur = fetch{lvalue: true, base: cur.value(), off: fld.ByteOffset, typ: fld.Type}
		case m[1] == ".":
			st, ok := t.(*dwarf.StructType)
			if !ok {
				return "", fmt.Errorf("%s: .%s 앞이 구조체가 아닙니다", expr, m[2])
			}
			fld := findField(st, m[2])
			if fld == nil {
				return "", fmt.Errorf("%s: %s에 %s가 없습니다", expr, st.StructName, m[2])
			}
			if !cur.lvalue {
				// 레지스터에 든 작은 구조체(반환값 FullTransactionId 등)는 첫 필드만 같은 레지스터로 읽을 수 있습니다.
				if fld.ByteOffset != 0 {
					return "", fmt.Errorf("%s: 레지스터 값의 0이 아닌 위치 필드는 읽을 수 없습니다", expr)
				}
				cur.typ = fld.Type
				continue
			}
			cur.off += fld.ByteOffset
			cur.typ = fld.Type
		default:
			n, _ := strconv.ParseInt(m[3], 10, 64)
			switch v := t.(type) {
			case *dwarf.PtrType:
				cur = fetch{lvalue: true, base: cur.value(), off: n * strip(v.Type).Size(), typ: v.Type}
			case *dwarf.ArrayType:
				if !cur.lvalue {
					return "", fmt.Errorf("%s: 메모리에 있지 않은 배열입니다", expr)
				}
				cur.off += n * strip(v.Type).Size()
				cur.typ = v.Type
			default:
				return "", fmt.Errorf("%s: [%d] 앞이 배열이나 포인터가 아닙니다", expr, n)
			}
		}
	}
	kind, err := fetchType(cur.typ)
	if err != nil {
		return "", fmt.Errorf("%s: %w", expr, err)
	}
	return cur.value() + ":" + kind, nil
}

func findField(st *dwarf.StructType, name string) *dwarf.StructField {
	for _, f := range st.Field {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// fetchType 은 DWARF 형식을 커널의 읽기 형식으로 바꿉니다. 포인터는 16진수(x64), 열거형은 크기만큼의 부호 있는 정수입니다.
func fetchType(t dwarf.Type) (string, error) {
	switch v := strip(t).(type) {
	case *dwarf.PtrType:
		return "x64", nil
	case *dwarf.BoolType:
		return "u8", nil
	case *dwarf.EnumType:
		return fmt.Sprintf("s%d", v.ByteSize*8), nil
	case *dwarf.IntType:
		return fmt.Sprintf("s%d", v.ByteSize*8), nil
	case *dwarf.UintType:
		return fmt.Sprintf("u%d", v.ByteSize*8), nil
	case *dwarf.CharType:
		return "s8", nil
	case *dwarf.UcharType:
		return "u8", nil
	case nil:
		return "", fmt.Errorf("형식을 모릅니다")
	default:
		return "", fmt.Errorf("읽을 수 없는 형식 %s(%T)", v, v)
	}
}

// parseProbes 는 관측 정의 파일을 읽습니다. "#"으로 시작하는 줄과 빈 줄은 건너뜁니다.
func parseProbes(path string) ([]probeDef, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []probeDef
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			return nil, fmt.Errorf("%s:%d: 이름과 함수가 필요합니다", path, n)
		}
		p := probeDef{Name: parts[0], Func: parts[1]}
		if fn, ok := strings.CutSuffix(p.Func, "%return"); ok {
			p.Func, p.Return = fn, true
		}
		for _, kv := range parts[2:] {
			k, v, ok := strings.Cut(kv, "=")
			if !ok {
				return nil, fmt.Errorf("%s:%d: %q는 필드=식 꼴이 아닙니다", path, n, kv)
			}
			p.Fields = append(p.Fields, [2]string{k, v})
		}
		out = append(out, p)
	}
	return out, sc.Err()
}

// uprobeLines 는 관측 정의마다 uprobe_events에 쓸 줄을 만듭니다. path는 OrbStack 머신에서 본 실행 파일 경로입니다.
//
// Errors: 함수를 찾지 못하거나 식 하나라도 풀지 못하면, 어느 정의의 어느 식인지 붙여 오류를 돌려줍니다.
func uprobeLines(bin, path string, defs []probeDef) ([]string, error) {
	want := map[string]bool{}
	for _, p := range defs {
		want[p.Func] = true
		for _, f := range p.Fields {
			if g := regexp.MustCompile(`^@(\w+)`).FindStringSubmatch(f[1]); g != nil {
				want["@"+g[1]] = true
			}
		}
	}
	idx, dup, err := loadDwarf(bin, want)
	if err != nil {
		return nil, err
	}
	for _, name := range dup {
		fmt.Fprintf(os.Stderr, "주의: 같은 이름의 함수가 여럿이라 처음 것을 씁니다: %s\n", name)
	}
	var out []string
	for _, p := range defs {
		fn, ok := idx.funcs[p.Func]
		if !ok {
			return nil, fmt.Errorf("%s: 함수 %s의 정의를 찾지 못했습니다", p.Name, p.Func)
		}
		kind := "p"
		if p.Return {
			kind = "r"
		}
		line := fmt.Sprintf("%s:txlab/%s %s:0x%x", kind, p.Name, path, fn.Low)
		for _, f := range p.Fields {
			arg, err := idx.resolve(fn, f[1])
			if err != nil {
				return nil, fmt.Errorf("%s: %w", p.Name, err)
			}
			line += " " + f[0] + "=" + arg
		}
		out = append(out, line)
	}
	return out, nil
}

// cmdUprobes 는 관측 정의를 uprobe 정의 줄로 바꿔 출력합니다. 기록 없이 정의만 확인할 때 씁니다.
func cmdUprobes(args []string) error {
	fs := flag.NewFlagSet("uprobes", flag.ContinueOnError)
	bin := fs.String("bin", "", "DWARF를 읽을 실행 파일")
	path := fs.String("path", "", "OrbStack 머신에서 본 실행 파일 경로(비우면 -bin)")
	defsPath := fs.String("probes", "probes.txt", "관측 정의 파일")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *path == "" {
		*path = *bin
	}
	defs, err := parseProbes(*defsPath)
	if err != nil {
		return err
	}
	lines, err := uprobeLines(*bin, *path, defs)
	if err != nil {
		return err
	}
	for _, l := range lines {
		fmt.Println(l)
	}
	return nil
}
