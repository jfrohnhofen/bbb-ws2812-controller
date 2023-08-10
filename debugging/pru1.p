.origin 0
.entrypoint start

#define gpio r30

#define SHIFT    9
#define STORE    8
#define DATA0    0
#define DATA1    1
#define TRIGGER 10

#define data0 r1
#define data1 r2
#define ones r3
#define zeros r4

.macro nop
.mparam n
    loop exit, n-1
    mov     r0, r0
exit:
.endm


.macro shift
.mparam data
    mov     gpio.b0, data
    nop     2
    set     gpio, SHIFT
    nop     2
    clr     gpio, SHIFT
    nop     2
.endm


.macro output
    shift   data0.b0
    shift   data0.b1
    shift   data0.b2
    shift   data0.b3
    shift   data1.b0
    shift   data1.b1
    shift   data1.b2
    shift   data1.b3
    set     gpio, STORE
    nop     2
    clr     gpio, STORE
.endm


start:
    mov     gpio, 0x00

    mov     ones, 0xffffffff
    mov     zeros, 0x00000000

    set     gpio, TRIGGER
    nop     100
    clr     gpio, TRIGGER
    
loop_bits:
    mov     data0, ones
    mov     data1, ones
    output
    nop     20

    mov     data0, 0x000000
    mov     data1, 0x000000
    output

    mov     data0, zeros
    mov     data1, zeros
    output

    nop  11

    jmp loop_bits
    
    halt