.origin 0
.entrypoint start

#define UINT32 4

#define CTRL_BASE_ADDR 0x22000
#define CTRL_CONTROL_REG 0x00
#define CTRL_CYCLE_REG 0x0c
#define CTRL_COUNTER_EN_BIT 3

#define tmp0    r28
#define tmp1    r29
#define gpio    r31

.macro cap4
.mparam rn
    mov     rn.b0, gpio.b0
    mov     rn.b1, gpio.b0
    mov     rn.b2, gpio.b0
    mov     rn.b3, gpio.b0
.endm

.macro capture
    cap4    r0
    cap4    r1
    cap4    r2
    cap4    r3
    cap4    r4
    cap4    r5
    cap4    r6
    cap4    r7
    cap4    r8
    cap4    r9
    cap4    r10
    cap4    r11
    cap4    r12
    cap4    r13
    cap4    r14
    cap4    r15
    cap4    r16
    cap4    r17
    cap4    r18
    cap4    r19
    cap4    r20
    cap4    r21
    cap4    r22
    cap4    r23
    cap4    r24
    cap4    r25
    cap4    r26
    cap4    r27
.endm

start:
    mov     tmp0, CTRL_BASE_ADDR
    lbbo    tmp1, tmp0, CTRL_CONTROL_REG, UINT32
    set     tmp1, CTRL_COUNTER_EN_BIT
    sbbo    tmp1, tmp0, CTRL_CONTROL_REG, UINT32

    wbs     gpio, 5

    lbbo    tmp1, tmp0, CTRL_CYCLE_REG, UINT32

    capture
    xout    10, r0, 28*UINT32
    capture
    xout    11, r0, 28*UINT32
    capture
    xout    12, r0, 28*UINT32
    capture

    lbbo    tmp0, tmp0, CTRL_CYCLE_REG, UINT32
    sub     tmp0, tmp0, tmp1
    
    mov     tmp1, 0x00
    sbbo    tmp0, tmp1, 0, UINT32

    mov     tmp1, 3*28*4 + 4
    sbbo    r0, tmp1, 0, 28*UINT32

    mov     tmp1, 2*28*4 + 4
    xin     12, r0, 28*UINT32
    sbbo    r0, tmp1, 0, 28*UINT32

    mov     tmp1, 1*28*4 + 4
    xin     11, r0, 28*UINT32
    sbbo    r0, tmp1, 0, 28*UINT32

    mov     tmp1, 0*28*4 + 4
    xin     10, r0, 28*UINT32
    sbbo    r0, tmp1, 0, 28*UINT32

    halt 
