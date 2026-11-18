.include "src_asm/include/axiom_defs.inc"
.include "src_asm/include/mca_memory.inc"
.include "src_asm/include/ast_defs.inc"

.equ TYPE_UNKNOWN, 0
.equ TYPE_INT,     1
.equ TYPE_STR,     2
.equ TYPE_ARR,     3
.equ TYPE_STRUCT,  4
.equ TYPE_STR_ARR, 6

.section .data
chk_kw_println: .ascii "println"
chk_kw_print:   .ascii "print"
chk_kw_len:     .ascii "len"
chk_kw_exit:    .ascii "exit"
chk_kw_argc:    .ascii "args_count"
chk_kw_arg:     .ascii "arg"
chk_kw_readf:   .ascii "read_file"
chk_kw_writef:  .ascii "write_file"
chk_kw_charat:  .ascii "char_at"
chk_kw_substr:  .ascii "substr"
chk_kw_toint:   .ascii "to_int"
chk_kw_spawn:   .ascii "__axiom_spawn"
chk_kw_fyield:  .ascii "fiber_yield"
chk_kw_frun:    .ascii "fiber_run_all"
chk_kw_frunpar: .ascii "fiber_run_parallel"
chk_kw_int2str: .ascii "int_to_str"
chk_kw_fdread:  .ascii "fd_read"
chk_kw_rerrmsg: .ascii "res_err_msg"
chk_kw_amark:   .ascii "arena_mark"
chk_kw_areset:  .ascii "arena_reset"
chk_kw_aused:   .ascii "arena_used"
chk_kw_mcreate: .ascii "map_create"
chk_kw_mset:    .ascii "map_set"
chk_kw_mget:    .ascii "map_get"
chk_kw_mhas:    .ascii "map_has"
chk_kw_rok:     .ascii "res_ok"
chk_kw_rerr:    .ascii "res_err"
chk_kw_risok:   .ascii "res_is_ok"
chk_kw_runwrap: .ascii "res_unwrap"
chk_kw_tcplis:  .ascii "tcp_listen"
chk_kw_tcpacc:  .ascii "tcp_accept"
chk_kw_tcpconn: .ascii "tcp_connect_local"
chk_kw_fdwrite: .ascii "fd_write"
chk_kw_fdclose: .ascii "fd_close"
chk_kw_sleepms: .ascii "sleep_ms"
chk_kw_timens:  .ascii "time_ns"
chk_kw_f64fi:   .ascii "f64_from_int"
chk_kw_f64ti:   .ascii "f64_to_int"
chk_kw_f64add:  .ascii "f64_add"
chk_kw_f64sub:  .ascii "f64_sub"
chk_kw_f64mul:  .ascii "f64_mul"
chk_kw_f64div:  .ascii "f64_div"
chk_kw_f64sqrt: .ascii "f64_sqrt"
chk_kw_f64lt:   .ascii "f64_lt"
chk_kw_f64ts:   .ascii "f64_to_str"
chk_kw_syscall: .ascii "sys_call"
.align 8
chk_beta_table:
    .quad chk_kw_int2str, 10, chk_kw_fdread, 7, chk_kw_rerrmsg, 11
    .quad chk_kw_amark, 10, chk_kw_areset, 11, chk_kw_aused, 10
    .quad chk_kw_mcreate, 10, chk_kw_mset, 7, chk_kw_mget, 7, chk_kw_mhas, 7
    .quad chk_kw_rok, 6, chk_kw_rerr, 7, chk_kw_risok, 9, chk_kw_runwrap, 10
    .quad chk_kw_tcplis, 10, chk_kw_tcpacc, 10, chk_kw_tcpconn, 17
    .quad chk_kw_fdwrite, 8, chk_kw_fdclose, 8, chk_kw_sleepms, 8, chk_kw_timens, 7
    .quad chk_kw_f64fi, 12, chk_kw_f64ti, 10, chk_kw_f64add, 7, chk_kw_f64sub, 7
    .quad chk_kw_f64mul, 7, chk_kw_f64div, 7, chk_kw_f64sqrt, 8, chk_kw_f64lt, 6
    .quad chk_kw_f64ts, 10, chk_kw_syscall, 8
    .quad 0, 0
chk_kw_tspawn:  .ascii "thread_spawn"
chk_kw_tjoin:   .ascii "thread_join"
chk_kw_ccreate: .ascii "chan_create"
chk_kw_csend:   .ascii "chan_send"
chk_kw_crecv:   .ascii "chan_recv"
chk_kw_acreate: .ascii "atomic_create"
chk_kw_aadd:    .ascii "atomic_add"
chk_kw_aload:   .ascii "atomic_load"

.section .bss
.align 8
.global chk_err_code, chk_err_line
chk_err_code:   .quad 0
chk_err_line:   .quad 0
chk_deny_io:    .quad 0
chk_deny_mem:   .quad 0

chk_fn_ptrs:    .skip 4096 * 8
chk_fn_lens:    .skip 4096 * 8
chk_fn_ret_types: .skip 4096 * 8
chk_fn_param_cnt: .skip 4096 * 8
chk_fn_params:  .skip 4096 * 8
chk_fn_cnt:     .quad 0

chk_var_ptrs:   .skip 4096 * 8
chk_var_lens:   .skip 4096 * 8
chk_var_types:  .skip 4096 * 8
chk_var_cnt:    .quad 0

.section .text
.global checker_verify_program
.global checker_validate_bounds
.global chk_get_fn_ret_type
.global chk_get_var_type

# ==============================================================================
# checker_verify_program(ast_root: %rdi) -> %rax (0 = VALID, 1 = INVARIANT_FAIL)
# ==============================================================================
checker_verify_program:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12

    testq %rdi, %rdi
    jz .L_chk_fail
    cmpq $AST_PROGRAM, AST_OFF_TYPE(%rdi)
    jne .L_chk_fail

    movq %rdi, %r12
    movq $0, chk_err_code(%rip)
    movq $0, chk_err_line(%rip)
    movq $0, chk_deny_io(%rip)
    movq $0, chk_deny_mem(%rip)
    movq $0, chk_fn_cnt(%rip)
    movq $0, chk_var_cnt(%rip)

    # Pass 1: Collect all fn names and require [...] capabilities
    movq AST_OFF_CHILD_PTR(%r12), %rdi
    call checker_collect_top

    # Pass 2: Full recursive semantic, type, capability & invariant check
    movq AST_OFF_CHILD_PTR(%r12), %rdi
    call checker_walk_list

    popq %r12
    leave
    ret

.L_chk_fail:
    movq $1, %rax
    popq %r12
    leave
    ret

checker_collect_top:
    pushq %r12
    movq %rdi, %r12
.L_cct_loop:
    testq %r12, %r12
    jz .L_cct_done
    movq AST_OFF_TYPE(%r12), %rax
    cmpq $AST_REQUIRE, %rax
    jne .L_cct_fn
    cmpq $0, AST_OFF_VAL(%r12)
    jne .L_cct_mem
    movq $1, chk_deny_io(%rip)
.L_cct_mem:
    movq AST_OFF_RIGHT_PTR(%r12), %rcx
    testq $0x10000, %rcx
    jz .L_cct_next
    testq $2, %rcx
    jnz .L_cct_next
    movq $1, chk_deny_mem(%rip)
    jmp .L_cct_next
.L_cct_fn:
    cmpq $AST_FN, %rax
    jne .L_cct_next
    movq chk_fn_cnt(%rip), %rcx
    cmpq $4096, %rcx
    jge .L_cct_next
    movq AST_OFF_NAME_PTR(%r12), %rax
    leaq chk_fn_ptrs(%rip), %rdx
    movq %rax, (%rdx, %rcx, 8)
    movq AST_OFF_NAME_LEN(%r12), %rax
    leaq chk_fn_lens(%rip), %rdx
    movq %rax, (%rdx, %rcx, 8)
    # Infer return type from first AST_RETURN in body
    pushq %rcx
    movq AST_OFF_CHILD_PTR(%r12), %rdi
    call chk_infer_fn_body_ret
    popq %rcx
    leaq chk_fn_ret_types(%rip), %rdx
    movq %rax, (%rdx, %rcx, 8)
    # Count parameters in AST_OFF_RIGHT_PTR(%r12)
    xorq %rax, %rax
    movq AST_OFF_RIGHT_PTR(%r12), %rdx
.L_cct_pcnt:
    testq %rdx, %rdx
    jz .L_cct_pdone
    incq %rax
    movq AST_OFF_NEXT_PTR(%rdx), %rdx
    jmp .L_cct_pcnt
.L_cct_pdone:
    leaq chk_fn_param_cnt(%rip), %rdx
    movq %rax, (%rdx, %rcx, 8)
    movq AST_OFF_RIGHT_PTR(%r12), %rax
    leaq chk_fn_params(%rip), %rdx
    movq %rax, (%rdx, %rcx, 8)
    incq chk_fn_cnt(%rip)
.L_cct_next:
    movq AST_OFF_NEXT_PTR(%r12), %r12
    jmp .L_cct_loop
.L_cct_done:
    popq %r12
    ret

checker_walk_list:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    movq %rdi, %r12
.L_wl_loop:
    testq %r12, %r12
    jz .L_wl_ok
    movq %r12, %rdi
    call checker_check_node
    testq %rax, %rax
    jnz .L_wl_fail
    movq AST_OFF_NEXT_PTR(%r12), %r12
    jmp .L_wl_loop
.L_wl_ok:
    xorq %rax, %rax
    popq %r12
    leave
    ret
.L_wl_fail:
    movq $1, %rax
    popq %r12
    leave
    ret

checker_check_node:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13
    movq %rdi, %r12

    movq AST_OFF_TYPE(%r12), %rax
    cmpq $1, %rax
    jl .L_cn_err_syntax
    cmpq $25, %rax
    jg .L_cn_err_syntax

    cmpq $AST_FN, %rax
    je .L_cn_fn
    cmpq $AST_ACTOR, %rax
    je .L_cn_actor_skip
    cmpq $AST_STRUCT, %rax
    je .L_cn_actor_skip
    cmpq $AST_REQUIRE, %rax
    je .L_cn_actor_skip
    cmpq $AST_CAP, %rax
    je .L_cn_actor_skip
    cmpq $AST_LET, %rax
    je .L_cn_let
    cmpq $AST_FOR, %rax
    je .L_cn_for
    cmpq $AST_ASSIGN, %rax
    je .L_cn_assign
    cmpq $AST_IDENT, %rax
    je .L_cn_ident
    cmpq $AST_CALL, %rax
    je .L_cn_call
    cmpq $AST_ARRAY, %rax
    je .L_cn_alloc_chk
    cmpq $AST_STRUCT_LIT, %rax
    je .L_cn_alloc_chk
    cmpq $AST_UNARY, %rax
    je .L_cn_unary
    cmpq $AST_EXPR_BIN, %rax
    je .L_cn_bin
    jmp .L_cn_children

.L_cn_actor_skip:
    xorq %rax, %rax
    popq %r13
    popq %r12
    leave
    ret

.L_cn_alloc_chk:
    cmpq $0, chk_deny_mem(%rip)
    jne .L_cn_err_mem
    jmp .L_cn_children

.L_cn_fn:
    movq chk_var_cnt(%rip), %r13
    movq $0, chk_var_cnt(%rip)

    movq AST_OFF_RIGHT_PTR(%r12), %rdi
.L_fn_p_loop:
    testq %rdi, %rdi
    jz .L_fn_body
    pushq %rdi
    movq AST_OFF_NAME_PTR(%rdi), %rax
    movq AST_OFF_NAME_LEN(%rdi), %rsi
    movq AST_OFF_VAL(%rdi), %rdx
    movq %rax, %rdi
    call chk_add_var
    popq %rdi
    movq AST_OFF_NEXT_PTR(%rdi), %rdi
    jmp .L_fn_p_loop

.L_fn_body:
    movq AST_OFF_CHILD_PTR(%r12), %rdi
    call checker_walk_list
    movq %r13, chk_var_cnt(%rip)
    popq %r13
    popq %r12
    leave
    ret

.L_cn_let:
    movq AST_OFF_CHILD_PTR(%r12), %rdi
    call checker_walk_list
    testq %rax, %rax
    jnz .L_cn_err

    movq AST_OFF_CHILD_PTR(%r12), %rdi
    call checker_infer_type
    movq %rax, %rdx
    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    call chk_add_var
    xorq %rax, %rax
    popq %r13
    popq %r12
    leave
    ret

.L_cn_for:
    movq chk_var_cnt(%rip), %r13
    movq AST_OFF_CHILD_PTR(%r12), %rdi
    call checker_walk_list
    testq %rax, %rax
    jnz .L_cn_err
    movq AST_OFF_CHILD_PTR(%r12), %rdi
    call checker_infer_type
    movq $TYPE_INT, %rdx
    cmpq $TYPE_STR_ARR, %rax
    jne .L_cnf_add
    movq $TYPE_STR, %rdx
.L_cnf_add:
    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    call chk_add_var
    movq AST_OFF_RIGHT_PTR(%r12), %rdi
    call checker_walk_list
    movq %r13, chk_var_cnt(%rip)
    popq %r13
    popq %r12
    leave
    ret

.L_cn_assign:
    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    call chk_has_var
    testq %rax, %rax
    jz .L_cn_err_var
    jmp .L_cn_children

.L_cn_ident:
    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    call chk_has_var
    testq %rax, %rax
    jz .L_cn_err_var
    xorq %rax, %rax
    popq %r13
    popq %r12
    leave
    ret

.L_cn_call:
    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_println(%rip), %rdx
    movq $7, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_io_call

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_print(%rip), %rdx
    movq $5, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_io_call

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_len(%rip), %rdx
    movq $3, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_exit(%rip), %rdx
    movq $4, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_argc(%rip), %rdx
    movq $10, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_arg(%rip), %rdx
    movq $3, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_readf(%rip), %rdx
    movq $9, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_io_call

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_writef(%rip), %rdx
    movq $10, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_io_call

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_charat(%rip), %rdx
    movq $7, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_substr(%rip), %rdx
    movq $6, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_toint(%rip), %rdx
    movq $6, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_fyield(%rip), %rdx
    movq $11, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_frun(%rip), %rdx
    movq $13, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_spawn(%rip), %rdx
    movq $13, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jz .L_cn_user_call
    movq AST_OFF_CHILD_PTR(%r12), %rcx
    testq %rcx, %rcx
    jz .L_cn_err_fn
    movq AST_OFF_NAME_PTR(%rcx), %rdi
    movq AST_OFF_NAME_LEN(%rcx), %rsi
    call chk_has_fn
    testq %rax, %rax
    jz .L_cn_err_fn
    movq AST_OFF_RIGHT_PTR(%r12), %rdi
    testq %rdi, %rdi
    jz .L_cn_actor_skip
    call checker_check_node
    jmp .L_cn_actor_skip

.L_cn_user_call:
    # Check concurrency/thread/channel/atomic builtins
    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_tspawn(%rip), %rdx
    movq $12, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jz .L_cn_chk_conc_others
    movq AST_OFF_CHILD_PTR(%r12), %rcx
    testq %rcx, %rcx
    jz .L_cn_err_fn
    movq AST_OFF_NAME_PTR(%rcx), %rdi
    movq AST_OFF_NAME_LEN(%rcx), %rsi
    call chk_has_fn
    testq %rax, %rax
    jz .L_cn_err_fn
    movq AST_OFF_CHILD_PTR(%r12), %rax
    movq AST_OFF_NEXT_PTR(%rax), %rdi
    testq %rdi, %rdi
    jz .L_cn_actor_skip
    call checker_check_node
    jmp .L_cn_actor_skip

.L_cn_chk_conc_others:
    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_tjoin(%rip), %rdx
    movq $11, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_ccreate(%rip), %rdx
    movq $11, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_csend(%rip), %rdx
    movq $9, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_crecv(%rip), %rdx
    movq $9, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_acreate(%rip), %rdx
    movq $13, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_aadd(%rip), %rdx
    movq $10, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_aload(%rip), %rdx
    movq $11, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_frunpar(%rip), %rdx
    movq $18, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children

    leaq chk_beta_table(%rip), %r13
.L_cn_beta_loop:
    movq (%r13), %rdx
    testq %rdx, %rdx
    jz .L_cn_beta_done
    movq 8(%r13), %rcx
    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cn_children
    addq $16, %r13
    jmp .L_cn_beta_loop
.L_cn_beta_done:

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    call chk_has_fn
    testq %rax, %rax
    jz .L_cn_err_fn
    # Verify argument count matches declared parameter count
    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    call chk_get_fn_param_cnt
    movq %rax, %rcx          # expected param count
    xorq %rax, %rax          # actual arg count
    movq AST_OFF_CHILD_PTR(%r12), %rdx
.L_cnc_acnt:
    testq %rdx, %rdx
    jz .L_cnc_achk
    incq %rax
    movq AST_OFF_NEXT_PTR(%rdx), %rdx
    jmp .L_cnc_acnt
.L_cnc_achk:
    cmpq %rcx, %rax
    jne .L_cn_err_type
    # Check argument types against declared parameter types
    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    call chk_get_fn_params
    movq %rax, %r13          # param node
    movq AST_OFF_CHILD_PTR(%r12), %rbx  # arg node
.L_cnc_tloop:
    testq %r13, %r13
    jz .L_cn_children
    testq %rbx, %rbx
    jz .L_cn_children
    movq AST_OFF_VAL(%r13), %rcx
    testq %rcx, %rcx
    jz .L_cnc_tnext
    pushq %rcx
    movq %rbx, %rdi
    call checker_infer_type
    popq %rcx
    testq %rax, %rax
    jz .L_cnc_tnext
    cmpq %rcx, %rax
    jne .L_cn_err_type
.L_cnc_tnext:
    movq AST_OFF_NEXT_PTR(%r13), %r13
    movq AST_OFF_NEXT_PTR(%rbx), %rbx
    jmp .L_cnc_tloop

.L_cn_io_call:
    cmpq $0, chk_deny_io(%rip)
    jne .L_cn_err_cap
    jmp .L_cn_children

.L_cn_unary:
    movq AST_OFF_CHILD_PTR(%r12), %rdi
    call checker_infer_type
    cmpq $TYPE_STR, %rax
    jge .L_cn_err_type
    jmp .L_cn_children

.L_cn_bin:
    cmpq $TOK_DOTDOT, AST_OFF_VAL(%r12)
    je .L_cn_children

    movq AST_OFF_CHILD_PTR(%r12), %rdi
    call checker_infer_type
    pushq %rax               # left type [0(%rsp)]
    movq AST_OFF_RIGHT_PTR(%r12), %rdi
    call checker_infer_type
    movq %rax, %rcx          # right type
    popq %rax                # left type

    # Check string + string, string == string, string != string
    movq AST_OFF_VAL(%r12), %rdx
    cmpq $TOK_PLUS, %rdx
    je .L_cnb_str_ok
    cmpq $TOK_EQ, %rdx
    je .L_cnb_str_ok
    cmpq $TOK_NEQ, %rdx
    je .L_cnb_str_ok
    jmp .L_cnb_num_only

.L_cnb_str_ok:
    cmpq $TYPE_STR, %rax
    jne .L_cnb_chk_r_str
    cmpq $TYPE_STR, %rcx
    je .L_cn_children
    cmpq $TYPE_UNKNOWN, %rcx
    je .L_cn_children
    jmp .L_cn_err_type
.L_cnb_chk_r_str:
    cmpq $TYPE_STR, %rcx
    jne .L_cnb_num_only
    cmpq $TYPE_UNKNOWN, %rax
    je .L_cn_children
    jmp .L_cn_err_type

.L_cnb_num_only:
    cmpq $TYPE_STR, %rax
    jge .L_cn_err_type
    cmpq $TYPE_STR, %rcx
    jge .L_cn_err_type

    cmpq $TOK_SLASH, AST_OFF_VAL(%r12)
    jne .L_cn_children
    movq AST_OFF_RIGHT_PTR(%r12), %rcx
    testq %rcx, %rcx
    jz .L_cn_err_syntax
    cmpq $AST_INT, AST_OFF_TYPE(%rcx)
    jne .L_cn_children
    cmpq $0, AST_OFF_VAL(%rcx)
    je .L_cn_err_div0

.L_cn_children:
    movq AST_OFF_CHILD_PTR(%r12), %rdi
    call checker_walk_list
    testq %rax, %rax
    jnz .L_cn_err

    movq AST_OFF_RIGHT_PTR(%r12), %rdi
    call checker_walk_list
    testq %rax, %rax
    jnz .L_cn_err

    xorq %rax, %rax
    popq %r13
    popq %r12
    leave
    ret

.L_cn_err_syntax: movq $1, %rax; jmp .L_cn_set_err
.L_cn_err_div0:   movq $2, %rax; jmp .L_cn_set_err
.L_cn_err_var:    movq $3, %rax; jmp .L_cn_set_err
.L_cn_err_fn:     movq $4, %rax; jmp .L_cn_set_err
.L_cn_err_cap:    movq $5, %rax; jmp .L_cn_set_err
.L_cn_err_type:   movq $6, %rax; jmp .L_cn_set_err
.L_cn_err_mem:    movq $7, %rax
.L_cn_set_err:
    cmpq $0, chk_err_code(%rip)
    jne .L_cn_err
    movq %rax, chk_err_code(%rip)
    movq AST_OFF_LINE(%r12), %rcx
    movq %rcx, chk_err_line(%rip)
.L_cn_err:
    movq $1, %rax
    popq %r13
    popq %r12
    leave
    ret

# ==============================================================================
# checker_infer_type(node: %rdi) -> %rax (TYPE_*)
# ==============================================================================
checker_infer_type:
    testq %rdi, %rdi
    jz .L_cit_unk
    movq AST_OFF_TYPE(%rdi), %rax
    cmpq $AST_INT, %rax
    je .L_cit_int
    cmpq $AST_EXPR_BIN, %rax
    je .L_cit_bin
    cmpq $AST_UNARY, %rax
    je .L_cit_int
    cmpq $AST_STRING, %rax
    je .L_cit_str
    cmpq $AST_ARRAY, %rax
    je .L_cit_arr
    cmpq $AST_STRUCT_LIT, %rax
    je .L_cit_struct
    cmpq $AST_CALL, %rax
    je .L_cit_call
    cmpq $AST_IDENT, %rax
    je .L_cit_ident
.L_cit_unk:
    movq $TYPE_UNKNOWN, %rax
    ret
.L_cit_call:
    pushq %r12
    movq %rdi, %r12
    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_arg(%rip), %rdx
    movq $3, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cit_call_str
    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_readf(%rip), %rdx
    movq $9, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cit_call_str

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_substr(%rip), %rdx
    movq $6, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cit_call_str

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_int2str(%rip), %rdx
    movq $10, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cit_call_str

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_fdread(%rip), %rdx
    movq $7, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cit_call_str

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_rerrmsg(%rip), %rdx
    movq $11, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cit_call_str

    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    leaq chk_kw_f64ts(%rip), %rdx
    movq $10, %rcx
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cit_call_str

    # Check user function return type
    movq AST_OFF_NAME_PTR(%r12), %rdi
    movq AST_OFF_NAME_LEN(%r12), %rsi
    call chk_get_fn_ret_type
    popq %r12
    ret
.L_cit_call_str:
    popq %r12
    movq $TYPE_STR, %rax
    ret

chk_infer_fn_body_ret:
    # %rdi = stmt list -> %rax = TYPE_*
    pushq %r12
    movq %rdi, %r12
.L_cifr_loop:
    testq %r12, %r12
    jz .L_cifr_int
    cmpq $AST_RETURN, AST_OFF_TYPE(%r12)
    je .L_cifr_hit
    movq AST_OFF_NEXT_PTR(%r12), %r12
    jmp .L_cifr_loop
.L_cifr_hit:
    movq AST_OFF_CHILD_PTR(%r12), %rdi
    call checker_infer_type
    popq %r12
    ret
.L_cifr_int:
    movq $TYPE_INT, %rax
    popq %r12
    ret

chk_get_fn_ret_type:
    pushq %r12
    pushq %r13
    pushq %r14
    movq %rdi, %r12
    movq %rsi, %r13
    xorq %r14, %r14
.L_cgfr_loop:
    cmpq chk_fn_cnt(%rip), %r14
    jge .L_cgfr_int
    leaq chk_fn_ptrs(%rip), %rax
    movq (%rax, %r14, 8), %rdx
    leaq chk_fn_lens(%rip), %rax
    movq (%rax, %r14, 8), %rcx
    movq %r12, %rdi
    movq %r13, %rsi
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cgfr_hit
    incq %r14
    jmp .L_cgfr_loop
.L_cgfr_hit:
    leaq chk_fn_ret_types(%rip), %rax
    movq (%rax, %r14, 8), %rax
    popq %r14
    popq %r13
    popq %r12
    ret
.L_cgfr_int:
    movq $TYPE_INT, %rax
    popq %r14
    popq %r13
    popq %r12
    ret

chk_get_fn_param_cnt:
    pushq %r12
    pushq %r13
    pushq %r14
    movq %rdi, %r12
    movq %rsi, %r13
    xorq %r14, %r14
.L_cgfp_loop:
    cmpq chk_fn_cnt(%rip), %r14
    jge .L_cgfp_zero
    leaq chk_fn_ptrs(%rip), %rax
    movq (%rax, %r14, 8), %rdx
    leaq chk_fn_lens(%rip), %rax
    movq (%rax, %r14, 8), %rcx
    movq %r12, %rdi
    movq %r13, %rsi
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cgfp_hit
    incq %r14
    jmp .L_cgfp_loop
.L_cgfp_hit:
    leaq chk_fn_param_cnt(%rip), %rax
    movq (%rax, %r14, 8), %rax
    popq %r14
    popq %r13
    popq %r12
    ret
.L_cgfp_zero:
    xorq %rax, %rax
    popq %r14
    popq %r13
    popq %r12
    ret

chk_get_fn_params:
    pushq %r12
    pushq %r13
    pushq %r14
    movq %rdi, %r12
    movq %rsi, %r13
    xorq %r14, %r14
.L_cgfpl_loop:
    cmpq chk_fn_cnt(%rip), %r14
    jge .L_cgfpl_zero
    leaq chk_fn_ptrs(%rip), %rax
    movq (%rax, %r14, 8), %rdx
    leaq chk_fn_lens(%rip), %rax
    movq (%rax, %r14, 8), %rcx
    movq %r12, %rdi
    movq %r13, %rsi
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cgfpl_hit
    incq %r14
    jmp .L_cgfpl_loop
.L_cgfpl_hit:
    leaq chk_fn_params(%rip), %rax
    movq (%rax, %r14, 8), %rax
    popq %r14
    popq %r13
    popq %r12
    ret
.L_cgfpl_zero:
    xorq %rax, %rax
    popq %r14
    popq %r13
    popq %r12
    ret
.L_cit_int:
    movq $TYPE_INT, %rax
    ret
.L_cit_bin:
    cmpq $TOK_PLUS, AST_OFF_VAL(%rdi)
    jne .L_cit_int
    movq AST_OFF_CHILD_PTR(%rdi), %rdi
    jmp checker_infer_type
.L_cit_str:
    movq $TYPE_STR, %rax
    ret
.L_cit_arr:
    movq AST_OFF_CHILD_PTR(%rdi), %rdi
    testq %rdi, %rdi
    jz .L_cit_arr_plain
    call checker_infer_type
    cmpq $TYPE_STR, %rax
    jne .L_cit_arr_plain
    movq $TYPE_STR_ARR, %rax
    ret
.L_cit_arr_plain:
    movq $TYPE_ARR, %rax
    ret
.L_cit_struct:
    movq $TYPE_STRUCT, %rax
    ret
.L_cit_ident:
    movq AST_OFF_NAME_LEN(%rdi), %rsi
    movq AST_OFF_NAME_PTR(%rdi), %rdi
    jmp chk_get_var_type

chk_add_var:
    # %rdi = ptr, %rsi = len, %rdx = type
    movq chk_var_cnt(%rip), %rcx
    cmpq $4096, %rcx
    jge .L_cav_done
    leaq chk_var_ptrs(%rip), %rax
    movq %rdi, (%rax, %rcx, 8)
    leaq chk_var_lens(%rip), %rax
    movq %rsi, (%rax, %rcx, 8)
    leaq chk_var_types(%rip), %rax
    movq %rdx, (%rax, %rcx, 8)
    incq chk_var_cnt(%rip)
.L_cav_done:
    ret

chk_get_var_type:
    pushq %r12
    pushq %r13
    pushq %r14
    movq %rdi, %r12
    movq %rsi, %r13
    movq chk_var_cnt(%rip), %r14
    decq %r14
.L_cgvt_loop:
    cmpq $0, %r14
    jl .L_cgvt_none
    leaq chk_var_ptrs(%rip), %rax
    movq (%rax, %r14, 8), %rdx
    leaq chk_var_lens(%rip), %rax
    movq (%rax, %r14, 8), %rcx
    movq %r12, %rdi
    movq %r13, %rsi
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_cgvt_hit
    decq %r14
    jmp .L_cgvt_loop
.L_cgvt_hit:
    leaq chk_var_types(%rip), %rax
    movq (%rax, %r14, 8), %rax
    popq %r14
    popq %r13
    popq %r12
    ret
.L_cgvt_none:
    xorq %rax, %rax
    popq %r14
    popq %r13
    popq %r12
    ret

chk_has_var:
    pushq %r12
    pushq %r13
    pushq %r14
    movq %rdi, %r12
    movq %rsi, %r13
    xorq %r14, %r14
.L_chv_loop:
    cmpq chk_var_cnt(%rip), %r14
    jge .L_chv_no
    leaq chk_var_ptrs(%rip), %rax
    movq (%rax, %r14, 8), %rdx
    leaq chk_var_lens(%rip), %rax
    movq (%rax, %r14, 8), %rcx
    movq %r12, %rdi
    movq %r13, %rsi
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_chv_yes
    incq %r14
    jmp .L_chv_loop
.L_chv_yes:
    movq $1, %rax
    popq %r14
    popq %r13
    popq %r12
    ret
.L_chv_no:
    xorq %rax, %rax
    popq %r14
    popq %r13
    popq %r12
    ret

chk_has_fn:
    pushq %r12
    pushq %r13
    pushq %r14
    movq %rdi, %r12
    movq %rsi, %r13
    xorq %r14, %r14
.L_chf_loop:
    cmpq chk_fn_cnt(%rip), %r14
    jge .L_chv_no
    leaq chk_fn_ptrs(%rip), %rax
    movq (%rax, %r14, 8), %rdx
    leaq chk_fn_lens(%rip), %rax
    movq (%rax, %r14, 8), %rcx
    movq %r12, %rdi
    movq %r13, %rsi
    call chk_slice_eq
    testq %rax, %rax
    jnz .L_chv_yes
    incq %r14
    jmp .L_chf_loop

chk_slice_eq:
    cmpq %rcx, %rsi
    jne .L_cse_ne
    xorq %r8, %r8
.L_cse_loop:
    cmpq %rsi, %r8
    jge .L_cse_eq
    movb (%rdi, %r8), %al
    cmpb (%rdx, %r8), %al
    jne .L_cse_ne
    incq %r8
    jmp .L_cse_loop
.L_cse_eq:
    movq $1, %rax
    ret
.L_cse_ne:
    xorq %rax, %rax
    ret

checker_validate_bounds:
    cmpq %rsi, %rdi
    jae bounds_violation
    xorq %rax, %rax
    ret

bounds_violation:
    movq $1, %rax
    ret
