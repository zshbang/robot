package calculator

type Stack struct {
	top  int			// 栈顶
	data []interface{}  // 栈内元素
}

func NewStack(size int) Stack {
	return Stack{
		top: -1,
		data: make([]interface{}, size),
	}
}

func (s *Stack)Peek() interface{}  {
	if s.IsEmpty() {
		return nil
	}
	return s.data[s.top]
}

func (s *Stack)Pop() interface{} {
	if s.IsEmpty() {
		return nil
	}
	e := s.data[s.top]
	s.top--
	return e
}

func (s *Stack)Push(e interface{}) {
	s.top++
	if s.top >= cap(s.data){
		s.data = append(s.data, e)
	} else {
		s.data[s.top] = e
	}
}

func (s *Stack) IsEmpty() bool {
	if s.top < 0 {
		return true
	} else {
		return false
	}
}

func (s *Stack) IsNotEmpty() bool {
	if s.top >= 0 {
		return true
	} else {
		return false
	}
}
