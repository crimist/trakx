package bencoding

import "github.com/crimist/trakx/internal/pool"

var dictionaryPool = pool.New(NewDictionary, func(dictionary *Dictionary) *Dictionary {
	dictionary.Reset()
	return dictionary
})

// AcquireDictionary returns a pooled dictionary ready for use.
func AcquireDictionary() *Dictionary {
	return dictionaryPool.Get()
}

// ReleaseDictionary returns a dictionary to the pool.
func ReleaseDictionary(dictionary *Dictionary) {
	if dictionary == nil {
		return
	}
	dictionaryPool.Put(dictionary)
}

// DictionaryPoolCreated returns the total number of dictionaries created by the pool.
func DictionaryPoolCreated() int64 {
	return dictionaryPool.Created()
}
