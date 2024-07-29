# Go tool objdump example1

- [Go tool objdump example1](#go-tool-objdump-example1)
    - [type assertion vs reflection](#type-assertion-vs-reflection)
        - [type assertion](#type-assertion)
        - [reflection](#reflection)
        - [분석](#분석)
            - [Type Assertion](#type-assertion-1)
            - [Reflection](#reflection-1)
            - [결론](#결론)

## type assertion vs reflection

### type assertion

```go
func main() {
    var val interface{} = complex128(100)

    switch val.(type) {
    case string:
        fmt.Println("string type")
    case int:
        fmt.Println("int type")
    case int16:
        fmt.Println("int16 type")
    case int32:
        fmt.Println("int32 type")
    case int64:
        fmt.Println("int64 type")
    case uint:
        fmt.Println("uint type")
    case uint8:
        fmt.Println("uint8 type")
    case uint16:
        fmt.Println("uint16 type")
    case uint32:
        fmt.Println("uint32 type")
    case uint64:
        fmt.Println("uint64 type")
    case float32:
        fmt.Println("float32 type")
    case float64:
        fmt.Println("float64 type")
    case complex64:
        fmt.Println("complex64 type")
    case complex128:
        fmt.Println("complex128 type")
    case bool:
        fmt.Println("bool type")
    default:
        fmt.Println("unknown type")
    }
}
```

```text
TEXT main.main(SB) /Users/rody/IdeaProjects/snippets/go/examples/control/switch/assertion/main.go
  main.go:5        0x10008b4e0        f9400b90        MOVD 16(R28), R16            
  main.go:5        0x10008b4e4        eb3063ff        CMP R16, RSP                
  main.go:5        0x10008b4e8        540007a9        BLS 61(PC)                
  main.go:5        0x10008b4ec        f81a0ffe        MOVD.W R30, -96(RSP)            
  main.go:5        0x10008b4f0        f81f83fd        MOVD R29, -8(RSP)            
  main.go:5        0x10008b4f4        d10023fd        SUB $8, RSP, R29            
  main.go:8        0x10008b4f8        90000145        ADRP 163840(PC), R5            
  main.go:8        0x10008b4fc        9106c0a5        ADD $432, R5, R5            
  main.go:8        0x10008b500        b94000a5        MOVWU (R5), R5                
  main.go:8        0x10008b504        d29a44a6        MOVD $53797, R6                
  main.go:8        0x10008b508        f2ae0746        MOVK $(28730<<16), R6            
  main.go:8        0x10008b50c        2b0600bf        CMNW R6, R5                
  main.go:8        0x10008b510        54000408        BHI 32(PC)                
  main.go:8        0x10008b514        d281a786        MOVD $3388, R6                
  main.go:8        0x10008b518        f2ac9fa6        MOVK $(25853<<16), R6            
  main.go:8        0x10008b51c        6b0600bf        CMPW R6, R5                
  main.go:8        0x10008b520        54000389        BLS 28(PC)                
  main.go:8        0x10008b524        d29e4186        MOVD $61964, R6                
  main.go:8        0x10008b528        f2aed106        MOVK $(30344<<16), R6            
  main.go:8        0x10008b52c        6b0600bf        CMPW R6, R5                
  main.go:8        0x10008b530        54000308        BHI 24(PC)                
  main.go:8        0x10008b534        d2879f47        MOVD $15610, R7                
  main.go:8        0x10008b538        f2ad7da7        MOVK $(27629<<16), R7            
  main.go:8        0x10008b53c        6b0700bf        CMPW R7, R5                
  main.go:8        0x10008b540        54000280        BEQ 20(PC)                
  main.go:8        0x10008b544        6b0600bf        CMPW R6, R5                
  main.go:8        0x10008b548        54000241        BNE 18(PC)                
  main.go:36        0x10008b54c        a904ffff        STP (ZR, ZR), 72(RSP)            
  main.go:36        0x10008b550        90000145        ADRP 163840(PC), R5            
  main.go:36        0x10008b554        913680a5        ADD $3488, R5, R5            
  main.go:36        0x10008b558        f90027e5        MOVD R5, 72(RSP)            
  main.go:36        0x10008b55c        d00001c5        ADRP 237568(PC), R5            
  main.go:36        0x10008b560        9125a0a5        ADD $2408, R5, R5            
  main.go:36        0x10008b564        f9002be5        MOVD R5, 80(RSP)            
  print.go:314        0x10008b568        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b56c        9133c0a5        ADD $3312, R5, R5            
  print.go:314        0x10008b570        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b574        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008b578        91376000        ADD $3544, R0, R0            
  print.go:314        0x10008b57c        910123e2        ADD $72, RSP, R2            
  print.go:314        0x10008b580        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008b584        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008b588        97ffeb02        CALL fmt.Fprintln(SB)            
  main.go:8        0x10008b58c        14000011        JMP 17(PC)                
  main.go:40        0x10008b590        a903ffff        STP (ZR, ZR), 56(RSP)            
  main.go:40        0x10008b594        90000145        ADRP 163840(PC), R5            
  main.go:40        0x10008b598        913680a5        ADD $3488, R5, R5            
  main.go:40        0x10008b59c        f9001fe5        MOVD R5, 56(RSP)            
  main.go:40        0x10008b5a0        d00001c5        ADRP 237568(PC), R5            
  main.go:40        0x10008b5a4        9125e0a5        ADD $2424, R5, R5            
  main.go:40        0x10008b5a8        f90023e5        MOVD R5, 64(RSP)            
  print.go:314        0x10008b5ac        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b5b0        9133c0a5        ADD $3312, R5, R5            
  print.go:314        0x10008b5b4        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b5b8        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008b5bc        91376000        ADD $3544, R0, R0            
  print.go:314        0x10008b5c0        9100e3e2        ADD $56, RSP, R2            
  print.go:314        0x10008b5c4        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008b5c8        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008b5cc        97ffeaf1        CALL fmt.Fprintln(SB)            
  main.go:42        0x10008b5d0        a97ffbfd        LDP -8(RSP), (R29, R30)            
  main.go:42        0x10008b5d4        910183ff        ADD $96, RSP, RSP            
  main.go:42        0x10008b5d8        d65f03c0        RET                    
  main.go:5        0x10008b5dc        aa1e03e3        MOVD R30, R3                
  main.go:5        0x10008b5e0        97ff3c7c        CALL runtime.morestack_noctxt.abi0(SB)    
  main.go:5        0x10008b5e4        17ffffbf        JMP main.main(SB)            
  main.go:5        0x10008b5e8        00000000        ?                    
  main.go:5        0x10008b5ec        00000000        ?                    
```

### reflection

```go
func main() {
    var val interface{} = complex128(100)

    switch reflect.ValueOf(val).Kind() {
    case reflect.String:
        fmt.Println("string type")
    case reflect.Int:
        fmt.Println("int type")
    case reflect.Int16:
        fmt.Println("int16 type")
    case reflect.Int32:
        fmt.Println("int32 type")
    case reflect.Int64:
        fmt.Println("int64 type")
    case reflect.Uint:
        fmt.Println("uint type")
    case reflect.Uint8:
        fmt.Println("uint8 type")
    case reflect.Uint16:
        fmt.Println("uint16 type")
    case reflect.Uint32:
        fmt.Println("uint32 type")
    case reflect.Uint64:
        fmt.Println("uint64 type")
    case reflect.Float32:
        fmt.Println("float32 type")
    case reflect.Float64:
        fmt.Println("float64 type")
    case reflect.Complex64:
        fmt.Println("complex64 type")
    case reflect.Complex128:
        fmt.Println("complex128 type")
    case reflect.Bool:
        fmt.Println("bool type")
    default:
        fmt.Println("unknown type")
    }
}
```

```text
TEXT main.main(SB) /Users/rody/IdeaProjects/snippets/go/examples/control/switch/reflection/main.go
  main.go:8        0x10008b4e0        f9400b90        MOVD 16(R28), R16            
  main.go:8        0x10008b4e4        d10343f1        SUB $208, RSP, R17            
  main.go:8        0x10008b4e8        eb10023f        CMP R16, R17                
  main.go:8        0x10008b4ec        540029a9        BLS 333(PC)                
  main.go:8        0x10008b4f0        d10543f4        SUB $336, RSP, R20            
  main.go:8        0x10008b4f4        a93ffa9d        STP (R29, R30), -8(R20)            
  main.go:8        0x10008b4f8        9100029f        MOVD R20, RSP                
  main.go:8        0x10008b4fc        d10023fd        SUB $8, RSP, R29            
  value.go:3148        0x10008b500        d503201f        NOOP                    
  main.go:11        0x10008b504        b00005a5        ADRP 741376(PC), R5            
  main.go:11        0x10008b508        910740a5        ADD $464, R5, R5            
  value.go:3852        0x10008b50c        394000a5        MOVBU (R5), R5                
  value.go:3852        0x10008b510        36000305        TBZ $0, R5, 24(PC)            
  value.go:3853        0x10008b514        90000145        ADRP 163840(PC), R5            
  value.go:3853        0x10008b518        910680a5        ADD $416, R5, R5            
  value.go:3853        0x10008b51c        b00005a6        ADRP 741376(PC), R6            
  value.go:3853        0x10008b520        910760c6        ADD $472, R6, R6            
  value.go:3853        0x10008b524        f90000c5        MOVD R5, (R6)                
  value.go:3853        0x10008b528        b0000746        ADRP 954368(PC), R6            
  value.go:3853        0x10008b52c        910380c6        ADD $224, R6, R6            
  value.go:3853        0x10008b530        b94000c6        MOVWU (R6), R6                
  value.go:3853        0x10008b534        350000e6        CBNZW R6, 7(PC)                
  value.go:3853        0x10008b538        f00000e6        ADRP 126976(PC), R6            
  value.go:3853        0x10008b53c        913240c6        ADD $3216, R6, R6            
  value.go:3853        0x10008b540        b00005a7        ADRP 741376(PC), R7            
  value.go:3853        0x10008b544        910780e7        ADD $480, R7, R7            
  value.go:3853        0x10008b548        f90000e6        MOVD R6, (R7)                
  value.go:3853        0x10008b54c        1400000d        JMP 13(PC)                
  value.go:3853        0x10008b550        b00005a2        ADRP 741376(PC), R2            
  value.go:3853        0x10008b554        91078042        ADD $480, R2, R2            
  value.go:3853        0x10008b558        f00000e3        ADRP 126976(PC), R3            
  value.go:3853        0x10008b55c        91324063        ADD $3216, R3, R3            
  value.go:3853        0x10008b560        97ff4584        CALL runtime.gcWriteBarrier(SB)        
  value.go:3853        0x10008b564        f00000e6        ADRP 126976(PC), R6            
  value.go:3853        0x10008b568        913240c6        ADD $3216, R6, R6            
  value.go:3853        0x10008b56c        14000005        JMP 5(PC)                
  value.go:3853        0x10008b570        90000145        ADRP 163840(PC), R5            
  value.go:3853        0x10008b574        910680a5        ADD $416, R5, R5            
  value.go:3853        0x10008b578        f00000e6        ADRP 126976(PC), R6            
  value.go:3853        0x10008b57c        913240c6        ADD $3216, R6, R6            
  value.go:3150        0x10008b580        f9001fe5        MOVD R5, 56(RSP)            
  value.go:3150        0x10008b584        f90023e6        MOVD R6, 64(RSP)            
  value.go:149        0x10008b588        f9401fe5        MOVD 56(RSP), R5            
  value.go:150        0x10008b58c        b4000125        CBZ R5, 9(PC)                
  type.go:823        0x10008b590        39405ca5        MOVBU 23(R5), R5            
  type.go:823        0x10008b594        924010a6        AND $31, R5, R6                
  value.go:155        0x10008b598        b27900c7        ORR $128, R6, R7            
  type.go:3127        0x10008b59c        927b00a5        AND $32, R5, R5                
  type.go:3127        0x10008b5a0        710000bf        CMPW $0, R5                
  value.go:157        0x10008b5a4        9a8600e5        CSEL EQ, R7, R6, R5            
  value.go:153        0x10008b5a8        d503201f        NOOP                    
  value.go:154        0x10008b5ac        14000002        JMP 2(PC)                
  value.go:154        0x10008b5b0        aa1f03e5        MOVD ZR, R5                
  value.go:85        0x10008b5b4        924010a5        AND $31, R5, R5                
  main.go:11        0x10008b5b8        d10004a5        SUB $1, R5, R5                
  value.go:1683        0x10008b5bc        d503201f        NOOP                    
  main.go:11        0x10008b5c0        f1005cbf        CMP $23, R5                
  main.go:11        0x10008b5c4        54002088        BHI 260(PC)                
  main.go:11        0x10008b5c8        f00001c0        ADRP 241664(PC), R0            
  main.go:11        0x10008b5cc        91198000        ADD $1632, R0, R0            
  main.go:11        0x10008b5d0        f865781b        MOVD (R0)(R5<<3), R27            
  main.go:11        0x10008b5d4        d61f0360        JMP (R27)                
  main.go:41        0x10008b5d8        a905ffff        STP (ZR, ZR), 88(RSP)            
  main.go:41        0x10008b5dc        90000145        ADRP 163840(PC), R5            
  main.go:41        0x10008b5e0        913680a5        ADD $3488, R5, R5            
  main.go:41        0x10008b5e4        f9002fe5        MOVD R5, 88(RSP)            
  main.go:41        0x10008b5e8        d00001c5        ADRP 237568(PC), R5            
  main.go:41        0x10008b5ec        912720a5        ADD $2504, R5, R5            
  main.go:41        0x10008b5f0        f90033e5        MOVD R5, 96(RSP)            
  print.go:314        0x10008b5f4        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b5f8        913440a5        ADD $3344, R5, R5            
  print.go:314        0x10008b5fc        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b600        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008b604        913ae000        ADD $3768, R0, R0            
  print.go:314        0x10008b608        910163e2        ADD $88, RSP, R2            
  print.go:314        0x10008b60c        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008b610        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008b614        97ffeadf        CALL fmt.Fprintln(SB)            
  main.go:41        0x10008b618        140000ff        JMP 255(PC)                
  main.go:15        0x10008b61c        a912ffff        STP (ZR, ZR), 296(RSP)            
  main.go:15        0x10008b620        90000145        ADRP 163840(PC), R5            
  main.go:15        0x10008b624        913680a5        ADD $3488, R5, R5            
  main.go:15        0x10008b628        f90097e5        MOVD R5, 296(RSP)            
  main.go:15        0x10008b62c        d00001c5        ADRP 237568(PC), R5            
  main.go:15        0x10008b630        9127a0a5        ADD $2536, R5, R5            
  main.go:15        0x10008b634        f9009be5        MOVD R5, 304(RSP)            
  print.go:314        0x10008b638        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b63c        913440a5        ADD $3344, R5, R5            
  print.go:314        0x10008b640        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b644        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008b648        913ae000        ADD $3768, R0, R0            
  print.go:314        0x10008b64c        9104a3e2        ADD $296, RSP, R2            
  print.go:314        0x10008b650        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008b654        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008b658        97ffeace        CALL fmt.Fprintln(SB)            
  main.go:15        0x10008b65c        140000ee        JMP 238(PC)                
  main.go:17        0x10008b660        a911ffff        STP (ZR, ZR), 280(RSP)            
  main.go:17        0x10008b664        90000145        ADRP 163840(PC), R5            
  main.go:17        0x10008b668        913680a5        ADD $3488, R5, R5            
  main.go:17        0x10008b66c        f9008fe5        MOVD R5, 280(RSP)            
  main.go:17        0x10008b670        d00001c5        ADRP 237568(PC), R5            
  main.go:17        0x10008b674        9127e0a5        ADD $2552, R5, R5            
  main.go:17        0x10008b678        f90093e5        MOVD R5, 288(RSP)            
  print.go:314        0x10008b67c        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b680        913440a5        ADD $3344, R5, R5            
  print.go:314        0x10008b684        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b688        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008b68c        913ae000        ADD $3768, R0, R0            
  print.go:314        0x10008b690        910463e2        ADD $280, RSP, R2            
  print.go:314        0x10008b694        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008b698        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008b69c        97ffeabd        CALL fmt.Fprintln(SB)            
  main.go:17        0x10008b6a0        140000dd        JMP 221(PC)                
  main.go:19        0x10008b6a4        a910ffff        STP (ZR, ZR), 264(RSP)            
  main.go:19        0x10008b6a8        90000145        ADRP 163840(PC), R5            
  main.go:19        0x10008b6ac        913680a5        ADD $3488, R5, R5            
  main.go:19        0x10008b6b0        f90087e5        MOVD R5, 264(RSP)            
  main.go:19        0x10008b6b4        d00001c5        ADRP 237568(PC), R5            
  main.go:19        0x10008b6b8        912820a5        ADD $2568, R5, R5            
  main.go:19        0x10008b6bc        f9008be5        MOVD R5, 272(RSP)            
  print.go:314        0x10008b6c0        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b6c4        913440a5        ADD $3344, R5, R5            
  print.go:314        0x10008b6c8        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b6cc        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008b6d0        913ae000        ADD $3768, R0, R0            
  print.go:314        0x10008b6d4        910423e2        ADD $264, RSP, R2            
  print.go:314        0x10008b6d8        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008b6dc        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008b6e0        97ffeaac        CALL fmt.Fprintln(SB)            
  main.go:19        0x10008b6e4        140000cc        JMP 204(PC)                
  main.go:21        0x10008b6e8        a90fffff        STP (ZR, ZR), 248(RSP)            
  main.go:21        0x10008b6ec        90000145        ADRP 163840(PC), R5            
  main.go:21        0x10008b6f0        913680a5        ADD $3488, R5, R5            
  main.go:21        0x10008b6f4        f9007fe5        MOVD R5, 248(RSP)            
  main.go:21        0x10008b6f8        d00001c5        ADRP 237568(PC), R5            
  main.go:21        0x10008b6fc        912860a5        ADD $2584, R5, R5            
  main.go:21        0x10008b700        f90083e5        MOVD R5, 256(RSP)            
  print.go:314        0x10008b704        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b708        913440a5        ADD $3344, R5, R5            
  print.go:314        0x10008b70c        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b710        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008b714        913ae000        ADD $3768, R0, R0            
  print.go:314        0x10008b718        9103e3e2        ADD $248, RSP, R2            
  print.go:314        0x10008b71c        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008b720        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008b724        97ffea9b        CALL fmt.Fprintln(SB)            
  main.go:21        0x10008b728        140000bb        JMP 187(PC)                
  main.go:23        0x10008b72c        a90effff        STP (ZR, ZR), 232(RSP)            
  main.go:23        0x10008b730        90000145        ADRP 163840(PC), R5            
  main.go:23        0x10008b734        913680a5        ADD $3488, R5, R5            
  main.go:23        0x10008b738        f90077e5        MOVD R5, 232(RSP)            
  main.go:23        0x10008b73c        d00001c5        ADRP 237568(PC), R5            
  main.go:23        0x10008b740        9128a0a5        ADD $2600, R5, R5            
  main.go:23        0x10008b744        f9007be5        MOVD R5, 240(RSP)            
  print.go:314        0x10008b748        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b74c        913440a5        ADD $3344, R5, R5            
  print.go:314        0x10008b750        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b754        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008b758        913ae000        ADD $3768, R0, R0            
  print.go:314        0x10008b75c        9103a3e2        ADD $232, RSP, R2            
  print.go:314        0x10008b760        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008b764        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008b768        97ffea8a        CALL fmt.Fprintln(SB)            
  main.go:23        0x10008b76c        140000aa        JMP 170(PC)                
  main.go:25        0x10008b770        a90dffff        STP (ZR, ZR), 216(RSP)            
  main.go:25        0x10008b774        90000145        ADRP 163840(PC), R5            
  main.go:25        0x10008b778        913680a5        ADD $3488, R5, R5            
  main.go:25        0x10008b77c        f9006fe5        MOVD R5, 216(RSP)            
  main.go:25        0x10008b780        d00001c5        ADRP 237568(PC), R5            
  main.go:25        0x10008b784        9128e0a5        ADD $2616, R5, R5            
  main.go:25        0x10008b788        f90073e5        MOVD R5, 224(RSP)            
  print.go:314        0x10008b78c        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b790        913440a5        ADD $3344, R5, R5            
  print.go:314        0x10008b794        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b798        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008b79c        913ae000        ADD $3768, R0, R0            
  print.go:314        0x10008b7a0        910363e2        ADD $216, RSP, R2            
  print.go:314        0x10008b7a4        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008b7a8        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008b7ac        97ffea79        CALL fmt.Fprintln(SB)            
  main.go:25        0x10008b7b0        14000099        JMP 153(PC)                
  main.go:27        0x10008b7b4        a90cffff        STP (ZR, ZR), 200(RSP)            
  main.go:27        0x10008b7b8        90000145        ADRP 163840(PC), R5            
  main.go:27        0x10008b7bc        913680a5        ADD $3488, R5, R5            
  main.go:27        0x10008b7c0        f90067e5        MOVD R5, 200(RSP)            
  main.go:27        0x10008b7c4        d00001c5        ADRP 237568(PC), R5            
  main.go:27        0x10008b7c8        912920a5        ADD $2632, R5, R5            
  main.go:27        0x10008b7cc        f9006be5        MOVD R5, 208(RSP)            
  print.go:314        0x10008b7d0        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b7d4        913440a5        ADD $3344, R5, R5            
  print.go:314        0x10008b7d8        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b7dc        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008b7e0        913ae000        ADD $3768, R0, R0            
  print.go:314        0x10008b7e4        910323e2        ADD $200, RSP, R2            
  print.go:314        0x10008b7e8        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008b7ec        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008b7f0        97ffea68        CALL fmt.Fprintln(SB)            
  main.go:27        0x10008b7f4        14000088        JMP 136(PC)                
  main.go:29        0x10008b7f8        a90bffff        STP (ZR, ZR), 184(RSP)            
  main.go:29        0x10008b7fc        90000145        ADRP 163840(PC), R5            
  main.go:29        0x10008b800        913680a5        ADD $3488, R5, R5            
  main.go:29        0x10008b804        f9005fe5        MOVD R5, 184(RSP)            
  main.go:29        0x10008b808        d00001c5        ADRP 237568(PC), R5            
  main.go:29        0x10008b80c        912960a5        ADD $2648, R5, R5            
  main.go:29        0x10008b810        f90063e5        MOVD R5, 192(RSP)            
  print.go:314        0x10008b814        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b818        913440a5        ADD $3344, R5, R5            
  print.go:314        0x10008b81c        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b820        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008b824        913ae000        ADD $3768, R0, R0            
  print.go:314        0x10008b828        9102e3e2        ADD $184, RSP, R2            
  print.go:314        0x10008b82c        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008b830        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008b834        97ffea57        CALL fmt.Fprintln(SB)            
  main.go:29        0x10008b838        14000077        JMP 119(PC)                
  main.go:31        0x10008b83c        a90affff        STP (ZR, ZR), 168(RSP)            
  main.go:31        0x10008b840        90000145        ADRP 163840(PC), R5            
  main.go:31        0x10008b844        913680a5        ADD $3488, R5, R5            
  main.go:31        0x10008b848        f90057e5        MOVD R5, 168(RSP)            
  main.go:31        0x10008b84c        d00001c5        ADRP 237568(PC), R5            
  main.go:31        0x10008b850        9125e0a5        ADD $2424, R5, R5            
  main.go:31        0x10008b854        f9005be5        MOVD R5, 176(RSP)            
  print.go:314        0x10008b858        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b85c        913440a5        ADD $3344, R5, R5            
  print.go:314        0x10008b860        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b864        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008b868        913ae000        ADD $3768, R0, R0            
  print.go:314        0x10008b86c        9102a3e2        ADD $168, RSP, R2            
  print.go:314        0x10008b870        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008b874        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008b878        97ffea46        CALL fmt.Fprintln(SB)            
  main.go:31        0x10008b87c        14000066        JMP 102(PC)                
  main.go:33        0x10008b880        a909ffff        STP (ZR, ZR), 152(RSP)            
  main.go:33        0x10008b884        90000145        ADRP 163840(PC), R5            
  main.go:33        0x10008b888        913680a5        ADD $3488, R5, R5            
  main.go:33        0x10008b88c        f9004fe5        MOVD R5, 152(RSP)            
  main.go:33        0x10008b890        d00001c5        ADRP 237568(PC), R5            
  main.go:33        0x10008b894        912620a5        ADD $2440, R5, R5            
  main.go:33        0x10008b898        f90053e5        MOVD R5, 160(RSP)            
  print.go:314        0x10008b89c        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b8a0        913440a5        ADD $3344, R5, R5            
  print.go:314        0x10008b8a4        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b8a8        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008b8ac        913ae000        ADD $3768, R0, R0            
  print.go:314        0x10008b8b0        910263e2        ADD $152, RSP, R2            
  print.go:314        0x10008b8b4        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008b8b8        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008b8bc        97ffea35        CALL fmt.Fprintln(SB)            
  main.go:33        0x10008b8c0        14000055        JMP 85(PC)                
  main.go:35        0x10008b8c4        a908ffff        STP (ZR, ZR), 136(RSP)            
  main.go:35        0x10008b8c8        90000145        ADRP 163840(PC), R5            
  main.go:35        0x10008b8cc        913680a5        ADD $3488, R5, R5            
  main.go:35        0x10008b8d0        f90047e5        MOVD R5, 136(RSP)            
  main.go:35        0x10008b8d4        d00001c5        ADRP 237568(PC), R5            
  main.go:35        0x10008b8d8        912660a5        ADD $2456, R5, R5            
  main.go:35        0x10008b8dc        f9004be5        MOVD R5, 144(RSP)            
  print.go:314        0x10008b8e0        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b8e4        913440a5        ADD $3344, R5, R5            
  print.go:314        0x10008b8e8        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b8ec        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008b8f0        913ae000        ADD $3768, R0, R0            
  print.go:314        0x10008b8f4        910223e2        ADD $136, RSP, R2            
  print.go:314        0x10008b8f8        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008b8fc        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008b900        97ffea24        CALL fmt.Fprintln(SB)            
  main.go:35        0x10008b904        14000044        JMP 68(PC)                
  main.go:37        0x10008b908        a907ffff        STP (ZR, ZR), 120(RSP)            
  main.go:37        0x10008b90c        90000145        ADRP 163840(PC), R5            
  main.go:37        0x10008b910        913680a5        ADD $3488, R5, R5            
  main.go:37        0x10008b914        f9003fe5        MOVD R5, 120(RSP)            
  main.go:37        0x10008b918        d00001c5        ADRP 237568(PC), R5            
  main.go:37        0x10008b91c        9126a0a5        ADD $2472, R5, R5            
  main.go:37        0x10008b920        f90043e5        MOVD R5, 128(RSP)            
  print.go:314        0x10008b924        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b928        913440a5        ADD $3344, R5, R5            
  print.go:314        0x10008b92c        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b930        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008b934        913ae000        ADD $3768, R0, R0            
  print.go:314        0x10008b938        9101e3e2        ADD $120, RSP, R2            
  print.go:314        0x10008b93c        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008b940        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008b944        97ffea13        CALL fmt.Fprintln(SB)            
  main.go:37        0x10008b948        14000033        JMP 51(PC)                
  main.go:39        0x10008b94c        a906ffff        STP (ZR, ZR), 104(RSP)            
  main.go:39        0x10008b950        90000145        ADRP 163840(PC), R5            
  main.go:39        0x10008b954        913680a5        ADD $3488, R5, R5            
  main.go:39        0x10008b958        f90037e5        MOVD R5, 104(RSP)            
  main.go:39        0x10008b95c        d00001c5        ADRP 237568(PC), R5            
  main.go:39        0x10008b960        9126e0a5        ADD $2488, R5, R5            
  main.go:39        0x10008b964        f9003be5        MOVD R5, 112(RSP)            
  print.go:314        0x10008b968        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b96c        913440a5        ADD $3344, R5, R5            
  print.go:314        0x10008b970        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b974        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008b978        913ae000        ADD $3768, R0, R0            
  print.go:314        0x10008b97c        9101a3e2        ADD $104, RSP, R2            
  print.go:314        0x10008b980        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008b984        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008b988        97ffea02        CALL fmt.Fprintln(SB)            
  main.go:39        0x10008b98c        14000022        JMP 34(PC)                
  main.go:13        0x10008b990        a913ffff        STP (ZR, ZR), 312(RSP)            
  main.go:13        0x10008b994        90000145        ADRP 163840(PC), R5            
  main.go:13        0x10008b998        913680a5        ADD $3488, R5, R5            
  main.go:13        0x10008b99c        f9009fe5        MOVD R5, 312(RSP)            
  main.go:13        0x10008b9a0        d00001c5        ADRP 237568(PC), R5            
  main.go:13        0x10008b9a4        9125a0a5        ADD $2408, R5, R5            
  main.go:13        0x10008b9a8        f900a3e5        MOVD R5, 320(RSP)            
  print.go:314        0x10008b9ac        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b9b0        913440a5        ADD $3344, R5, R5            
  print.go:314        0x10008b9b4        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b9b8        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008b9bc        913ae000        ADD $3768, R0, R0            
  print.go:314        0x10008b9c0        9104e3e2        ADD $312, RSP, R2            
  print.go:314        0x10008b9c4        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008b9c8        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008b9cc        97ffe9f1        CALL fmt.Fprintln(SB)            
  main.go:13        0x10008b9d0        14000011        JMP 17(PC)                
  main.go:43        0x10008b9d4        a904ffff        STP (ZR, ZR), 72(RSP)            
  main.go:43        0x10008b9d8        90000145        ADRP 163840(PC), R5            
  main.go:43        0x10008b9dc        913680a5        ADD $3488, R5, R5            
  main.go:43        0x10008b9e0        f90027e5        MOVD R5, 72(RSP)            
  main.go:43        0x10008b9e4        d00001c5        ADRP 237568(PC), R5            
  main.go:43        0x10008b9e8        912760a5        ADD $2520, R5, R5            
  main.go:43        0x10008b9ec        f9002be5        MOVD R5, 80(RSP)            
  print.go:314        0x10008b9f0        900005a5        ADRP 737280(PC), R5            
  print.go:314        0x10008b9f4        913440a5        ADD $3344, R5, R5            
  print.go:314        0x10008b9f8        f94000a1        MOVD (R5), R1                
  print.go:314        0x10008b9fc        d00001c0        ADRP 237568(PC), R0            
  print.go:314        0x10008ba00        913ae000        ADD $3768, R0, R0            
  print.go:314        0x10008ba04        910123e2        ADD $72, RSP, R2            
  print.go:314        0x10008ba08        b24003e3        ORR $1, ZR, R3                
  print.go:314        0x10008ba0c        aa0303e4        MOVD R3, R4                
  print.go:314        0x10008ba10        97ffe9e0        CALL fmt.Fprintln(SB)            
  main.go:45        0x10008ba14        a97ffbfd        LDP -8(RSP), (R29, R30)            
  main.go:45        0x10008ba18        910543ff        ADD $336, RSP, RSP            
  main.go:45        0x10008ba1c        d65f03c0        RET                    
  main.go:8        0x10008ba20        aa1e03e3        MOVD R30, R3                
  main.go:8        0x10008ba24        97ff3b6b        CALL runtime.morestack_noctxt.abi0(SB)    
  main.go:8        0x10008ba28        17fffeae        JMP main.main(SB)            
  main.go:8        0x10008ba2c        00000000        ?                    
```

### 분석

Type assertion과 reflection을 사용하여 타입을 파악하는 두 코드의 objdump 출력을 비교해보면, 몇 가지 주요 차이점을 발견할 수 있습니다. 이러한 차이점들은 각 방법의 효율성과 성능에 영향을 미칩니다.

#### Type Assertion

- Type assertion 방식은 주어진 인터페이스 값의 실제 타입을 직접 확인합니다. 이 과정에서 컴파일러는 타입 정보를 직접 사용하여 런타임에 타입 체크를 수행합니다.
- Objdump 출력에서 볼 수 있듯이, type assertion 코드는 상대적으로 간단한 명령어 시퀀스를 가지고 있습니다. 이는 타입 확인 과정이 직접적이고, 필요한 타입 정보가 컴파일 시간에 이미 결정되기 때문입니다.
- Type assertion은 주로 `CMP`, `BLS`, `BEQ`, `BNE` 등의 명령어를 사용하여 타입을 확인하고, 조건에 따라 분기합니다. 이는 상대적으로 CPU 사이클이 적게 소모되는 작업입니다.

#### Reflection

- Reflection 방식은 `reflect` 패키지의 기능을 사용하여 런타임에 객체의 타입 정보를 검사합니다. 이 과정은 더 복잡하고, 내부적으로 더 많은 작업을 수행합니다.
- Objdump 출력에서 reflection 코드는 type assertion 코드보다 훨씬 더 많은 명령어를 포함하고 있습니다. 이는 reflection이 메모리에서 타입 정보를 조회하고, 여러 단계를 거쳐 타입을 확인하기 때문입니다.
- Reflection을 사용하는 코드는 `MOVD`, `SUB`, `STP`, `MOVBU`, `CBNZW` 등 다양한 명령어를 사용합니다. 이는 reflection이 메모리 접근, 조건 분기, 함수 호출 등 복잡한 작업을 포함하기 때문입니다.

#### 결론

- **효율성과 성능**: Type assertion은 reflection에 비해 더 직접적이고 간단한 방식으로 타입을 확인하기 때문에, 일반적으로 더 효율적이고 빠릅니다. Reflection은 더 많은 CPU 사이클과 메모리 접근을 필요로 하며, 이로 인해 성능이 더 낮아질 수 있습니다.
- **사용 사례**: Type assertion은 컴파일 시간에 가능한 타입이 이미 알려져 있고, 런타임에 이를 확인하기만 하면 되는 경우에 적합합니다. 반면, reflection은 런타임에 타입 정보를 동적으로 검사하고 조작해야 할 때 유용합니다. 이는 더 유연하지만, 성능 비용이 더 높습니다.
- **결론**: 성능이 중요한 상황에서는 가능한 한 type assertion을 사용하는 것이 좋습니다. 하지만, 코드의 유연성과 동적 타입 처리가 필요한 경우에는 reflection을 사용할 수 있습니다. 각각의 방법은 사용 사례와 성능 요구 사항에 따라 선택해야 합니다.
