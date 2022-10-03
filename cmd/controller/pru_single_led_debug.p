.origin 0
.entrypoint start

#define DATA_PIN 10
#define TRIGGER_PIN 11
#define NUM_BITS 24
#define DATA 0b101010101111111100000000
#define HIGH_CYCLES 40
#define DATA_CYCLES 40
#define LOW_CYCLES 45

start:
    set     r30, r30, TRIGGER_PIN
    mov     r0, DATA
    mov     r1, NUM_BITS

loop_bits:
    set     r30, r30, DATA_PIN
    mov     r2, HIGH_CYCLES

wait_high:
    sub     r2, r2, 1
    qbne    wait_high, r2, 0

    mov     r2, DATA_CYCLES
    qbbs    wait_data, r0, 0
    clr     r30, r30, DATA_PIN

wait_data:
    sub     r2, r2, 1
    qbne    wait_data, r2, 0

    clr     r30, r30, DATA_PIN
    mov     r2, LOW_CYCLES

wait_low:
    sub     r2, r2, 1
    qbne    wait_low, r2, 0

    lsr     r0, r0, 1
    sub     r1, r1, 1
    qbne    loop_bits, r1, 0

    clr     r30, r30, TRIGGER_PIN
    halt
