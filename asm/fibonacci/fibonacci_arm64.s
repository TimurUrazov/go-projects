#include "textflag.h"

// func Fibonacci(n uint64) uint64
TEXT ·Fibonacci(SB), NOSPLIT, $0
    MOVD n+0(FP), R0
    MOVD $0, R1
    MOVD $1, R2
loop:
    CBZ R0, return
    MOVD R2, R3
    ADD R1, R2
    MOVD R3, R1
    SUB $1, R0
    B loop

return:
    MOVD R1, ret+8(FP)
    RET
