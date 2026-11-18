# ==============================================================================
# AxiomLang Pure Freestanding HTTP/1.1 & CoreNet Network Client
# Direct Linux sys_socket, sys_connect, sys_write, sys_read
# Zero External Dependencies | Zero Libc
# ==============================================================================

.include "src_asm/include/axiom_defs.inc"

.equ SYS_SOCKET,  41
.equ SYS_CONNECT, 42
.equ AF_INET,     2
.equ SOCK_STREAM, 1

.section .rodata
http_get_prefix:
    .ascii "GET "
http_get_prefix_len = . - http_get_prefix

http_get_suffix:
    .ascii " HTTP/1.1\r\nHost: localhost\r\nUser-Agent: AxiomBrowser/1.0 (Freestanding x86_64)\r\nConnection: close\r\n\r\n"
http_get_suffix_len = . - http_get_suffix

msg_net_connected:
    .ascii "[CoreNet] TCP Socket successfully established.\n"
msg_net_connected_len = . - msg_net_connected

msg_net_req_sent:
    .ascii "[CoreNet] HTTP Request transmitted.\n"
msg_net_req_sent_len = . - msg_net_req_sent

msg_net_read_done:
    .ascii "[CoreNet] HTTP Response received via Zero-Copy stream.\n"
msg_net_read_done_len = . - msg_net_read_done

.section .bss
.align 16
net_sockfd:        .quad 0
.global net_response_buf, net_response_len
net_response_buf:  .skip 65536         # 64 KB response buffer
net_response_len:  .quad 0

.section .text
.global axiom_net_http_get

# ==============================================================================
# axiom_net_http_get(ip_u32: %rdi, port_u16: %rsi, path_ptr: %rdx, path_len: %rcx) -> %rax (bytes read)
# Performs a direct HTTP GET request to ip:port
# ==============================================================================
axiom_net_http_get:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13
    pushq %r14
    pushq %r15

    movq %rdi, %r12                       # IP (network byte order or 127.0.0.1 = 0x0100007f)
    movq %rsi, %r13                       # Port
    movq %rdx, %r14                       # path ptr
    movq %rcx, %r15                       # path len

    # 1. socket(AF_INET, SOCK_STREAM, 0)
    movq $SYS_SOCKET, %rax
    movq $AF_INET, %rdi
    movq $SOCK_STREAM, %rsi
    xorq %rdx, %rdx
    syscall
    testq %rax, %rax
    js .L_net_err
    movq %rax, net_sockfd(%rip)

    # 2. connect(sockfd, &sockaddr_in, 16)
    # sockaddr_in: sin_family=2 (AF_INET), sin_port (big endian), sin_addr (big endian)
    subq $16, %rsp
    movw $AF_INET, (%rsp)

    # convert port to big-endian (e.g. 8089 = 0x1f99 -> 0x991f)
    movq %r13, %rax
    rolw $8, %ax
    movw %ax, 2(%rsp)

    # store IP
    movl %r12d, 4(%rsp)
    movq $0, 8(%rsp)                      # sin_zero

    movq $SYS_CONNECT, %rax
    movq net_sockfd(%rip), %rdi
    movq %rsp, %rsi
    movq $16, %rdx
    syscall
    addq $16, %rsp
    testq %rax, %rax
    js .L_net_err

    # Print connected message
    movq $SYS_WRITE, %rax
    movq $1, %rdi
    leaq msg_net_connected(%rip), %rsi
    movq $msg_net_connected_len, %rdx
    syscall

    # 3. Send "GET " + path + " HTTP/1.1\r\n..."
    # Send GET prefix
    movq $SYS_WRITE, %rax
    movq net_sockfd(%rip), %rdi
    leaq http_get_prefix(%rip), %rsi
    movq $http_get_prefix_len, %rdx
    syscall

    # Send path
    movq $SYS_WRITE, %rax
    movq net_sockfd(%rip), %rdi
    movq %r14, %rsi
    movq %r15, %rdx
    syscall

    # Send suffix
    movq $SYS_WRITE, %rax
    movq net_sockfd(%rip), %rdi
    leaq http_get_suffix(%rip), %rsi
    movq $http_get_suffix_len, %rdx
    syscall

    # 4. Read response into net_response_buf
    xorq %r14, %r14                       # total bytes read
.L_read_loop:
    movq $SYS_READ, %rax
    movq net_sockfd(%rip), %rdi
    leaq net_response_buf(%rip), %rsi
    addq %r14, %rsi
    movq $65536, %rdx
    subq %r14, %rdx
    jle .L_read_done
    syscall

    testq %rax, %rax
    jle .L_read_done
    addq %rax, %r14
    jmp .L_read_loop

.L_read_done:
    movq %r14, net_response_len(%rip)

    # Close socket
    movq $SYS_CLOSE, %rax
    movq net_sockfd(%rip), %rdi
    syscall

    movq %r14, %rax
    popq %r15
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

.L_net_err:
    movq $0, %rax
    popq %r15
    popq %r14
    popq %r13
    popq %r12
    leave
    ret
