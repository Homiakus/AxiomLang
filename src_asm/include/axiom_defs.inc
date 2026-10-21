# ==============================================================================
# AxiomLang Pure x86_64 Freestanding Assembly Compiler Core
# System call numbers and common constants
# ==============================================================================

.equ SYS_READ,   0
.equ SYS_WRITE,  1
.equ SYS_OPEN,   2
.equ SYS_CLOSE,  3
.equ SYS_STAT,   4
.equ SYS_FSTAT,  5
.equ SYS_MMAP,   9
.equ SYS_MUNMAP, 11
.equ SYS_EXIT,   60
.equ SYS_UNLINK, 87

.equ O_RDONLY,   0
.equ O_WRONLY,   1
.equ O_CREAT,    64
.equ O_TRUNC,    512

.equ PROT_READ,  1
.equ PROT_WRITE, 2
.equ MAP_PRIVATE, 2
.equ MAP_ANONYMOUS, 32

.equ STDIN_FILENO,  0
.equ STDOUT_FILENO, 1
.equ STDERR_FILENO, 2

# Token IDs
.equ TOK_EOF,      0
.equ TOK_ILLEGAL,  1
.equ TOK_IDENT,    2
.equ TOK_INT,      3
.equ TOK_STRING,   4
.equ TOK_REQUIRE,  5
.equ TOK_CAP,      6
.equ TOK_INTENT,   7
.equ TOK_LET,      8
.equ TOK_MUT,      9
.equ TOK_FN,       10
.equ TOK_RETURN,   11
.equ TOK_IF,       12
.equ TOK_ELSE,     13
.equ TOK_FOR,      14
.equ TOK_IN,       15
.equ TOK_TRUE,     16
.equ TOK_FALSE,    17
.equ TOK_STRUCT,   18
.equ TOK_ACTOR,    19
.equ TOK_SPAWN,    20
.equ TOK_SEND,     21
.equ TOK_RECEIVE,  22

# Delimiters & Operators
.equ TOK_ASSIGN,   30  # =
.equ TOK_EQ,       31  # ==
.equ TOK_ARROW,    32  # =>
.equ TOK_PLUS,     33  # +
.equ TOK_MINUS,    34  # -
.equ TOK_ASTERISK, 35  # *
.equ TOK_SLASH,    36  # /
.equ TOK_BANG,     37  # !
.equ TOK_LT,       38  # <
.equ TOK_GT,       39  # >
.equ TOK_LPAREN,   40  # (
.equ TOK_RPAREN,   41  # )
.equ TOK_LBRACE,   42  # {
.equ TOK_RBRACE,   43  # }
.equ TOK_LBRACKET, 44  # [
.equ TOK_RBRACKET, 45  # ]
.equ TOK_COMMA,    46  # ,
.equ TOK_COLON,    47  # :
.equ TOK_SEMICOLON,48  # ;
