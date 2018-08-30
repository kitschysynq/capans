package capans

import "bytes"

var msgs = &MsgB0rker{}

type MsgB0rker struct {
	messages []*bytes.Buffer
}

func (m *MsgB0rker) Stream() *Stream {
	s := &Stream{
		id: len(m.messages),
		m:  m,
	}
	m.messages = append(m.messages, &bytes.Buffer{})
	return s
}

func (m *MsgB0rker) WriteFrom(id int, p []byte) (int, error) {
	for i := range m.messages {
		if i == id {
			continue
		}
		_, err := m.messages[i].Write(p)
		if err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func (m *MsgB0rker) ReadFor(id int, p []byte) (int, error) {
	return m.messages[id].Read(p)
}

type Stream struct {
	id int
	m  *MsgB0rker
}

func (s *Stream) Read(p []byte) (int, error) {
	return s.m.ReadFor(s.id, p)
}

func (s *Stream) Write(p []byte) (int, error) {
	return s.m.WriteFrom(s.id, p)
}

func (s *Stream) Id() int {
	return s.id
}
