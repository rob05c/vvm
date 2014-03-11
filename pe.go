package main

type ProcessingElement struct {
	ArithmeticRegister int64
	RoutingRegister    int64
	Index              int64
	Enabled            bool
	Memory             []int64

	Lod  chan ByteTuple
	Sto  chan ByteTuple
	Add  chan ByteTuple
	Sub  chan ByteTuple
	Mul  chan ByteTuple
	Div  chan ByteTuple
	Mov  chan ByteTuple
	Radd chan bool
	Rsub chan bool
	Rmul chan bool
	Rdiv chan bool

	Done chan bool ///< the PE writes to this when an instruction finishes.
}

func (pe *ProcessingElement) Run() {
	for {
		select {
		case p := <-pe.Lod:
			pe.DoLod(p.First, p.Second)
		case p := <-pe.Sto:
			pe.DoSto(p.First, p.Second)
		case p := <-pe.Add:
			pe.DoAdd(p.First, p.Second)
		case p := <-pe.Sub:
			pe.DoSub(p.First, p.Second)
		case p := <-pe.Mul:
			pe.DoMul(p.First, p.Second)
		case p := <-pe.Div:
			pe.DoDiv(p.First, p.Second)
		case p := <-pe.Mov:
			pe.DoMov(RegisterType(p.First), RegisterType(p.Second))
		case <-pe.Radd:
			pe.DoRadd()
		case <-pe.Rsub:
			pe.DoRsub()
		case <-pe.Rmul:
			pe.DoRmul()
		case <-pe.Rdiv:
			pe.DoRdiv()
		}
		pe.Done <- true
	}
}

///
/// PE (vector) instructions
///
func (pe *ProcessingElement) DoAdd(a byte, i byte) {
	if !pe.Enabled {
		return
	}
	pe.ArithmeticRegister += pe.Memory[a+i]
}
func (pe *ProcessingElement) DoSub(a byte, i byte) {
	if !pe.Enabled {
		return
	}
	pe.ArithmeticRegister -= pe.Memory[a+i]
}
func (pe *ProcessingElement) DoMul(a byte, i byte) {
	if !pe.Enabled {
		return
	}
	pe.ArithmeticRegister *= pe.Memory[a+i]
}
func (pe *ProcessingElement) DoDiv(a byte, i byte) {
	if !pe.Enabled {
		return
	}
	if pe.ArithmeticRegister == 0 {
		return
	}
	pe.ArithmeticRegister /= pe.Memory[a+i]
}

// lod operation for individual PE
// @todo change this to be signalled by a channel
func (pe *ProcessingElement) DoLod(a byte, i byte) {
	if !pe.Enabled {
		return
	}
	pe.ArithmeticRegister = pe.Memory[a+i]

}

// sto operation for individual PE
// @todo change this to be signalled by a channel
func (pe *ProcessingElement) DoSto(a byte, i byte) {
	if !pe.Enabled {
		return
	}
	pe.Memory[a+i] = pe.ArithmeticRegister
}

// @todo make this more efficient
func (pe *ProcessingElement) DoMov(from RegisterType, to RegisterType) {
	if !pe.Enabled {
		return
	}
	var fromVal int64
	switch from {
	case peIndex:
		fromVal = pe.Index
	case peRouting:
		fromVal = pe.RoutingRegister
	case peArithmetic:
		fromVal = pe.ArithmeticRegister
	}
	switch to {
	case peIndex:
		pe.Index = fromVal
	case peRouting:
		pe.RoutingRegister = fromVal
	case peArithmetic:
		pe.ArithmeticRegister = fromVal
	}
}
func (pe *ProcessingElement) DoRadd() {
	if !pe.Enabled {
		return
	}
	pe.ArithmeticRegister += pe.RoutingRegister
}
func (pe *ProcessingElement) DoRsub() {
	if !pe.Enabled {
		return
	}
	pe.ArithmeticRegister -= pe.RoutingRegister
}
func (pe *ProcessingElement) DoRmul() {
	if !pe.Enabled {
		return
	}
	pe.ArithmeticRegister *= pe.RoutingRegister
}
func (pe *ProcessingElement) DoRdiv() {
	if !pe.Enabled {
		return
	}
	pe.ArithmeticRegister /= pe.RoutingRegister
}
