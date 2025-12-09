package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

// tire is set by InitTrie; nil means not initialized and TrieReplace returns the input unchanged.
var tire *Trie

var trieReplaceNotInitOnce sync.Once

const defaultTrieConfigPath = "./configs/trie.txt"

// InitTrie loads sensitive words from defaultTrieConfigPath into the trie used by TrieReplace.
// If the file does not exist, logs a warning and returns nil (trie is empty).
func InitTrie() error {
	tr := NewTrie()
	tire = &tr
	f, err := os.Open(defaultTrieConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warnf("trie config not found: %s", defaultTrieConfigPath)
			return nil
		}
		return err
	}
	defer f.Close()

	buff := bufio.NewReader(f)
	for {
		line, _, err := buff.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read trie config: %w", err)
		}
		tire.Inster(string(line))
	}
	return nil
}

// TrieReplace masks sensitive substrings. Call InitTrie first; if InitTrie was never run, returns str unchanged.
func TrieReplace(str string) string {
	if tire == nil {
		trieReplaceNotInitOnce.Do(func() {
			log.Warn("TrieReplace: InitTrie was not called; returning input unchanged")
		})
		return str
	}
	return tire.Replace(str)
}
