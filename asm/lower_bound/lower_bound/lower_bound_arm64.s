#include "textflag.h"

// func LowerBound(slice []int64, value int64) int64
TEXT ·LowerBound(SB), NOSPLIT, $0
    LDP slice_base+0(FP), (R0, R1)
    MOVD value+24(FP), R2
    MOVD $-1, R3
loop:
    SUB R3, R1, R4
    CMP $1, R4
    BLE return
    ADD R1, R3, R5
    LSR $1, R5, R5
    LSL $3, R5, R6
    ADD R0, R6
    MOVD (R6), R7
    CMP R2, R7
    BLE L1
    B L0
L0:
    MOVD R5, R1
    B loop
L1:
    MOVD R5, R3
    B loop
return:
    SUB $1, R1, R1
    MOVD R1, ret+32(FP)
    RET
