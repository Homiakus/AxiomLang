# ==============================================================================
# AxiomLang Mandate Capability Arena (MCA) System Definition
# Zero Garbage Collection, Deterministic O(1) Allocations, Memory Safety
# ==============================================================================

# Mandate Memory Model (MCA) Constants
.equ MCA_PAGE_SIZE,       4096
.equ MCA_DEFAULT_ARENA,   65536     # 64 KB default per-actor/mandate arena
.equ MCA_MAGIC,           0x41584D43  # "AXMC" in ASCII

# Memory Capability Bitmask
.equ CAP_MEM_READ,        1
.equ CAP_MEM_WRITE,       2
.equ CAP_MEM_EXEC,        4
.equ CAP_MEM_GROW,        8

# MCA Header Structure offset
.equ MCA_OFF_MAGIC,       0
.equ MCA_OFF_CAPS,        4
.equ MCA_OFF_TOTAL_SIZE,  8
.equ MCA_OFF_USED_SIZE,   16
.equ MCA_OFF_PARENT_PTR,  24
.equ MCA_HEADER_SIZE,     32
