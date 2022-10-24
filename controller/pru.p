.origin 0
.entrypoint start

#define UINT32 4

#define NUM_CHANNELS 6
#define RESET_CYCLES 200000

#define CTRL_BASE_ADDR 0x24000
#define CTRL_CONTROL_REG 0x00
#define CTRL_CYCLE_REG 0x0c
#define CTRL_COUNTER_EN_BIT 3

// frame data
#define END_ADDR_OFFSET 0
#define NEXT_ADDR_OFFSET 4
#define DATA_OFFSET 8

// pins
#define STORE_CLOCK_PIN 8
#define SHIFT_CLOCK_PIN 9
#define TRIGGER_PIN 11

// registers
#define tmp r0
#define buf_addr r1
#define ctrl r2
#define data0 r3
#define data1 r4
#define end_addr r6
#define data_addr r9
#define cnt0 r7
#define cnt1 r8
#define zeros r10
#define ones r11
#define gpio r30

.macro nop
.mparam n
    loop exit, n-1
    mov     tmp, tmp
exit:
.endm

.macro shift
.mparam data
    mov     gpio.b0, data
    nop     2
    set     gpio, SHIFT_CLOCK_PIN
    nop     2
    clr     gpio, SHIFT_CLOCK_PIN
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
    set     gpio, STORE_CLOCK_PIN
    nop     2
    clr     gpio, STORE_CLOCK_PIN
.endm


start:
    // constants
    mov     zeros, 0x00000000
    mov     ones, 0xffffffff

    // clear pins
    //mov     gpio, zeros

    // enable cycle counter
    mov     ctrl, CTRL_BASE_ADDR
    lbbo    tmp, ctrl, CTRL_CONTROL_REG, UINT32
    set     tmp, CTRL_COUNTER_EN_BIT
    sbbo    tmp, ctrl, CTRL_CONTROL_REG, UINT32

    // set address and params of first buffer
    mov     buf_addr, 0x00000000
    lbbo    end_addr, buf_addr, END_ADDR_OFFSET, UINT32
    add     data_addr, buf_addr, DATA_OFFSET

    set     gpio, TRIGGER_PIN

loop_bits:
    // output ones
    mov     data0, ones
    mov     data1, ones
    output
    nop     2

    // output data
    lbbo    data0, data_addr, 0, NUM_CHANNELS
    output
    
    // output zeros
    mov     data0, zeros
    mov     data1, zeros
    output
    
    // check exit condition
    qbeq    next_buffer, data_addr, end_addr
    add     data_addr, data_addr, NUM_CHANNELS
    
    nop 9
    jmp     loop_bits

next_buffer:
    // clear end_addr
    sbbo    zeros, buf_addr, END_ADDR_OFFSET, UINT32

    // load address of next buffer
    lbbo    buf_addr, buf_addr, NEXT_ADDR_OFFSET, UINT32

    // check exit condition
    qbeq    exit, buf_addr, ones

    // set params of next buffer
    lbbo    end_addr, buf_addr, END_ADDR_OFFSET, UINT32
    add     data_addr, buf_addr, DATA_OFFSET

    jmp     loop_bits

exit:
    nop 4

    mov     tmp, RESET_CYCLES
reset:
    sub     tmp, tmp, 2
    qbne    reset, tmp, 0

    clr     gpio, TRIGGER_PIN
    halt
