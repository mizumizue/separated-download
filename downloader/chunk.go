package downloader

type ChunkCollection []Chunk

type Chunk []byte

func (cc ChunkCollection) Join() []byte {
	res := make([]byte, 0)
	for _, c := range cc {
		res = append(res, c...)
	}
	return res
}
