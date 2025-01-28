#include "textflag.h"

// func WordCount(data []rune) int32
TEXT ·WordCount(SB), NOSPLIT, $0
    LDP slice_base+0(FP), (R0, R1)
    MOVW $0, R2
    MOVW $0, R3
loop:
    CBZ R1, return
    MOVW (R0), R4
    CMP $0xA0, R4
    BEQ check_flag
    CMP $0x20, R4
    BEQ check_flag
    CMP $0x85, R4
    BEQ check_flag
    CMP $0x2028, R4
    BEQ check_flag
    CMP $0x2029, R4
    BEQ check_flag
    CMP $0x202F, R4
    BEQ check_flag
    CMP $0x205F, R4
    BEQ check_flag
    CMP $0x3000, R4
    BEQ check_flag
    CMP $0x1680, R4
    BEQ check_flag
    CMP $0x09, R4
    BLT flag_incr
    CMP $0x0d, R4
    BGT check_en_quad_to_hair_space
    B check_flag
check_en_quad_to_hair_space:
    CMP $0x2000, R4
    BLT flag_incr
    CMP $0x200A, R4
    BGT flag_incr
check_flag:
    CMP $0, R3
    BEQ skip
ascend:
    ADD $1, R2
    MOVW $0, R3
skip:
    ADD $4, R0
    SUB $1, R1
    B loop
flag_incr:
    MOVW $1, R3
    B skip
return:
    ADD R3, R2
    MOVD R2, ret+24(FP)
    RET
