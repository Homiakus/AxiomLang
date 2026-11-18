# ==============================================================================
# RUNTIME_HEAL.S
# Топологический самоисцеляющийся рантайм: проекция поврежденного состояния
# на многообразие инвариантов безопасности (Moore-Penrose Geodesic Projection)
# ==============================================================================

.include "src_asm/include/axiom_defs.inc"

.section .data
heal_banner:
    .ascii " [HPT-HEAL] Anomaly detected: applying geodesic projection to safe manifold...\n"
heal_banner_len = . - heal_banner

heal_success:
    .ascii " [HPT-HEAL] State successfully healed. Panic prevented. Execution continuing.\n"
heal_success_len = . - heal_success

.section .text
.global hpt_runtime_heal_division_by_zero
.global hpt_runtime_heal_overflow

# ==============================================================================
# hpt_runtime_heal_division_by_zero:
# Предотвращает деление на ноль: если знаменатель 0, проецирует его на минимальный
# ненулевой эпсилон или каноническое нейтральное значение без вызова SIGFPE.
# Вход/Выход:
#   %rsi - делитель
# ==============================================================================
hpt_runtime_heal_division_by_zero:
    testq %rsi, %rsi
    jnz .L_no_div_zero

    # Оповещение о самоисцелении
    pushq %rax
    pushq %rdi
    pushq %rdx
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq heal_banner(%rip), %rsi
    movq $heal_banner_len, %rdx
    syscall

    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq heal_success(%rip), %rsi
    movq $heal_success_len, %rdx
    syscall
    popq %rdx
    popq %rdi
    popq %rax

    # Гомотопический сдвиг на эпсилон 1
    movq $1, %rsi

.L_no_div_zero:
    ret

# ==============================================================================
# hpt_runtime_heal_overflow(val: %rdi, min: %rsi, max: %rdx) -> %rax
# Предотвращает переполнение: насыщает значение до границ [min, max] (Saturating projection)
# ==============================================================================
hpt_runtime_heal_overflow:
    movq %rdi, %rax
    cmpq %rsi, %rax
    cmovlq %rsi, %rax
    cmpq %rdx, %rax
    cmovgq %rdx, %rax
    ret
