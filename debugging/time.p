#define CTRL_CONTROL_REG 0x00
#define CTRL_CYCLE_REG 0x0c
#define CTRL_COUNTER_EN_BIT 3

#define UINT32 4

// registers
#define time_ctrl r10
#define time_start r11
#define time_end r12
#define time_count r13
#define time_sum r14

.macro t_init
    mov     time_ctrl, CTRL_BASE_ADDR
    lbbo    time_start, time_ctrl, CTRL_CONTROL_REG, UINT32
    set     time_start, CTRL_COUNTER_EN_BIT
    sbbo    time_start, time_ctrl, CTRL_CONTROL_REG, UINT32
    mov     time_count, 0x00
    mov     time_sum, 0x00
.endm

.macro t_fini
    mov     time_ctrl, 0x00
    sbbo    time_count, time_ctrl, 0, UINT32*2
.endm

.macro t_start
    lbbo    time_start, time_ctrl, CTRL_CYCLE_REG, UINT32
.endm

.macro t_end
    lbbo    time_end, time_ctrl, CTRL_CYCLE_REG, UINT32
    sub     time_end, time_end, time_start
    add     time_count, time_count, 1
    add     time_sum, time_sum, time_end
.endm
