package capans

import (
	"bytes"
	"io"
)

var msgs = &MsgB0rker{}

type MsgB0rker struct {
	messages []*bytes.Buffer
	free     []int
}

func (m *MsgB0rker) Stream() *Stream {
	if len(m.free) == 0 {
		id := len(m.messages)
		s := &Stream{
			id: id,
			m:  m,
		}
		m.messages = append(m.messages, &bytes.Buffer{})
		s.buf = m.messages[id]
		return s
	}
	id := m.free[0]
	m.free = m.free[1:]
	s := &Stream{
		id: id,
		m:  m,
	}
	m.messages[id] = &bytes.Buffer{}
	s.buf = m.messages[id]
	return s
}

func (m *MsgB0rker) WriteFrom(id int, p []byte) (int, error) {
	for i := range m.messages {
		if i == id || m.messages[i] == nil {
			continue
		}
		_, err := m.messages[i].Write(p)
		if err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func (m *MsgB0rker) Close(id int) {
	m.messages[id] = nil
	m.free = append(m.free, id)
}

type Stream struct {
	id  int
	m   *MsgB0rker
	buf io.Reader
}

func (s *Stream) Id() int {
	return s.id
}

func (s *Stream) Read(p []byte) (int, error)  { return s.buf.Read(p) }
func (s *Stream) Write(p []byte) (int, error) { return s.m.WriteFrom(s.id, p) }
func (s *Stream) Close()                      { s.m.Close(s.id) }
