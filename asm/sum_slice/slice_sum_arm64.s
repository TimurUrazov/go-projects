#include "textflag.h"

// func SumSlice(s []int32) int64
TEXT ·SumSlice(SB), NOSPLIT, $0
    LDP slice_base+0(FP), (R0, R1)
    MOVD $0, R2
    MOVD $0, R3
loop:
    CBZ R1, return
    ADD R0, R2, R4
    MOVW (R4), R5
    ADD R5, R3
    SUB $1, R1
    ADD $4, R2
    B loop
return:
    MOVD R3, ret+24(FP)
    RET
