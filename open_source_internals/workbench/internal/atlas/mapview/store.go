package mapview

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"study/open_source_internals/workbench/internal/atlas"
)

// Store 는 보기 모형이 기록을 읽는 유일한 자리입니다.
//
// 저장 방식은 T2가 정하므로, 화면은 이 인터페이스로만 run을 받습니다. 다른 저장(DB, 내용 주소 저장)을 고르면 구현만 더합니다.
type Store interface {
	// List는 읽을 수 있는 run id를 이름 차례로 돌려줍니다.
	List() ([]string, error)
	// Load는 run 하나를 읽습니다. 없거나 형식이 틀리면 오류입니다.
	Load(id string) (*atlas.Run, error)
}

// DirStore 값은 공통 형식 JSON 파일을 <Dir>/<run id>.json으로 둔 폴더입니다(주제의 적재 명령이 쓰는 꼴).
type DirStore struct {
	Dir string
}

// List 메서드는 폴더 아래의 .json 파일을 모두 찾아 run id(폴더 기준 경로에서 .json을 뺀 것)를 이름 차례로 돌려줍니다.
func (d DirStore) List() ([]string, error) {
	var ids []string
	err := filepath.WalkDir(d.Dir, func(p string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() || !strings.HasSuffix(p, ".json") {
			return nil
		}
		rel, err := filepath.Rel(d.Dir, p)
		if err != nil {
			return err
		}
		ids = append(ids, filepath.ToSlash(strings.TrimSuffix(rel, ".json")))
		return nil
	})
	sort.Strings(ids)
	return ids, err
}

// Load 메서드는 run 하나를 형식대로 읽습니다(atlas.ReadFile, 형식에 없는 칸이 있으면 오류).
func (d DirStore) Load(id string) (*atlas.Run, error) {
	return atlas.ReadFile(filepath.Join(d.Dir, filepath.FromSlash(id)+".json"))
}

// MemStore 값은 이미 읽은 run을 id마다 둔 것입니다. 시험과, 적재 결과를 파일로 쓰지 않고 바로 보이는 쓰임에 씁니다.
type MemStore map[string]*atlas.Run

// List 메서드는 run id를 이름 차례로 돌려줍니다.
func (m MemStore) List() ([]string, error) {
	ids := make([]string, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids, nil
}

// Load 메서드는 id의 run을 돌려줍니다. 없으면 os.ErrNotExist를 감싼 오류입니다.
func (m MemStore) Load(id string) (*atlas.Run, error) {
	r, ok := m[id]
	if !ok {
		return nil, &fs.PathError{Op: "load", Path: id, Err: os.ErrNotExist}
	}
	return r, nil
}
