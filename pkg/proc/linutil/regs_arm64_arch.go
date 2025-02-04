package linutil

import (
	"fmt"
	"golang.org/x/arch/arm64/arm64asm"

	"github.com/go-delve/delve/pkg/proc"
)

// Regs is a wrapper for sys.PtraceRegs.
type ARM64Registers struct {
	Regs     *ARM64PtraceRegs //通用寄存器
	Fpregs   []proc.Register  //格式化后的浮点寄存器（名称，值，显示方式）
	Fpregset []byte           //保存所有浮点寄存器值的字节数组
}

// ARM64PtraceRegs is the struct used by the linux kernel to return the
// general purpose registers for ARM64 CPUs.
//copy form sys/unix/ztypes_linux_arm64.go:735
type ARM64PtraceRegs struct {
	Regs   [31]uint64
	Sp     uint64
	Pc     uint64
	Pstate uint64
}

// Slice returns the registers as a list of (name, value) pairs.
func (r *ARM64Registers) Slice(floatingPoint bool) []proc.Register {
	var regs64 = []struct {
		k string
		v uint64
	}{
		{"X0", r.Regs.Regs[0]},
		{"X1", r.Regs.Regs[1]},
		{"X2", r.Regs.Regs[2]},
		{"X3", r.Regs.Regs[3]},
		{"X4", r.Regs.Regs[4]},
		{"X5", r.Regs.Regs[5]},
		{"X6", r.Regs.Regs[6]},
		{"X7", r.Regs.Regs[7]},
		{"X8", r.Regs.Regs[8]},
		{"X9", r.Regs.Regs[9]},
		{"X10", r.Regs.Regs[10]},
		{"X11", r.Regs.Regs[11]},
		{"X12", r.Regs.Regs[12]},
		{"X13", r.Regs.Regs[13]},
		{"X14", r.Regs.Regs[14]},
		{"X15", r.Regs.Regs[15]},
		{"X16", r.Regs.Regs[16]},
		{"X17", r.Regs.Regs[17]},
		{"X18", r.Regs.Regs[18]},
		{"X19", r.Regs.Regs[19]},
		{"X20", r.Regs.Regs[20]},
		{"X21", r.Regs.Regs[21]},
		{"X22", r.Regs.Regs[22]},
		{"X23", r.Regs.Regs[23]},
		{"X24", r.Regs.Regs[24]},
		{"X25", r.Regs.Regs[25]},
		{"X26", r.Regs.Regs[26]},
		{"X27", r.Regs.Regs[27]},
		{"X28", r.Regs.Regs[28]},
		{"X29", r.Regs.Regs[29]},
		{"X30", r.Regs.Regs[30]},
		{"SP", r.Regs.Sp},
		{"PC", r.Regs.Pc},
		{"PSTATE", r.Regs.Pstate},
	}
	out := make([]proc.Register, 0, len(regs64)+len(r.Fpregs))
	for _, reg := range regs64 {
		out = proc.AppendQwordReg(out, reg.k, reg.v)
	}
	out = append(out, r.Fpregs...)
	return out
}

// PC returns the value of RIP register.
func (r *ARM64Registers) PC() uint64 {
	return r.Regs.Pc
}

// SP returns the value of RSP register.
func (r *ARM64Registers) SP() uint64 {
	return r.Regs.Sp
}

func (r *ARM64Registers) BP() uint64 {
	return r.Regs.Regs[29]
}

// CX returns the value of RCX register.
func (r *ARM64Registers) CX() uint64 {
	return 0
}

// TLS returns the address of the thread local storage memory segment.
func (r *ARM64Registers) TLS() uint64 {
	return 0
}

// GAddr returns the address of the G variable if it is known, 0 and false
// otherwise.
func (r *ARM64Registers) GAddr() (uint64, bool) {
	return r.Regs.Regs[28], true
}

// Get returns the value of the n-th register (in x86asm order).
func (r *ARM64Registers) Get(n int) (uint64, error) {
	reg := arm64asm.Reg(n)

	if reg >= arm64asm.X0 && reg <= arm64asm.X30 {
		return r.Regs.Regs[reg-arm64asm.X0], nil
	}

	return 0, proc.ErrUnknownRegister
}

// Copy returns a copy of these registers that is guarenteed not to change.
func (r *ARM64Registers) Copy() proc.Registers {
	var rr ARM64Registers
	rr.Regs = &ARM64PtraceRegs{}
	*(rr.Regs) = *(r.Regs)
	if r.Fpregs != nil {
		rr.Fpregs = make([]proc.Register, len(r.Fpregs))
		copy(rr.Fpregs, r.Fpregs)
	}
	if r.Fpregset != nil {
		rr.Fpregset = make([]byte, len(r.Fpregset))
		copy(rr.Fpregset, r.Fpregset)
	}
	return &rr
}

// Decode decodes an XSAVE area to a list of name/value pairs of registers.
func Decode(fpregs []byte) (regs []proc.Register) {
	for i := 0; i < len(fpregs); i += 16 {
		regs = proc.AppendFPReg(regs, fmt.Sprintf("V%d", i/16), fpregs[i:i+16])
	}
	return
}
