TEXT ·Prefetch(SB), $0
    MOVQ dPtr+0(FP), AX
    PREFETCHT0 (AX)
    RET
    