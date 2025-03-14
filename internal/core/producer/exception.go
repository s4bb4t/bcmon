package producer

func (p *Producer) Exception(contract string) {
	p.Lock()
	defer p.Unlock()
	p.exceptions[contract] = struct{}{}
}

func (p *Producer) excepted(contract string) bool {
	p.RLock()
	defer p.RUnlock()

	_, exists := p.exceptions[contract]

	return exists
}
