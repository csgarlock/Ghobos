#include "textflag.h"

TEXT ·Prefetch(SB), NOSPLIT, $0-8
    MOVQ dPtr+0(FP), AX
    PREFETCHT0 (AX)
    RET
