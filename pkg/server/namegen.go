package server

import (
	"fmt"
	"math/rand"
	"slices"
	"sync"
)

type NameGen struct {
	adj_idx         int
	adj_offset      int
	noun_idx        int
	noun_offset_idx int
	noun_offsets    []int

	mutex sync.Mutex
	names []string
}

func InitNameGen() *NameGen {
	noun_offsets := make([]int, len(nouns))
	for i := range len(nouns) {
		noun_offsets[i] = i
	}

	for i := len(noun_offsets) - 1; i > 0; i-- {
		j := rand.Intn(len(noun_offsets))
		temp := noun_offsets[i]
		noun_offsets[i] = noun_offsets[j]
		noun_offsets[j] = temp
	}
	fmt.Println(noun_offsets)

	return &NameGen{
		adj_idx:         0,
		adj_offset:      rand.Intn(len(adjectives)),
		noun_idx:        0,
		noun_offset_idx: 0,
		noun_offsets:    noun_offsets,
	}
}

func (ng *NameGen) newName() string {
	fmt.Println((ng.noun_idx + ng.noun_offsets[ng.noun_offset_idx]) % len(nouns))
	adj := adjectives[(ng.adj_idx+ng.adj_offset)%len(adjectives)]
	noun := nouns[(ng.noun_idx+ng.noun_offsets[ng.noun_offset_idx])%len(nouns)]

	ng.adj_idx += 1
	ng.adj_idx %= len(adjectives)
	ng.noun_idx += 1
	ng.noun_idx %= len(nouns)
	if ng.adj_idx == 0 {
		ng.noun_idx = 0
		ng.noun_offset_idx += 1
		ng.noun_offset_idx %= len(ng.noun_offsets)
	}

	return adj + " " + noun
}

func (ng *NameGen) insert(name string) bool {
	ng.mutex.Lock()
	if slices.Contains(ng.names, name) {
		fmt.Println("Nope")
		return false
	}
	ng.names = append(ng.names, name)
	ng.mutex.Unlock()
	return true
}

var adjectives = [100]string{
	"Happy", "Sad", "Angry", "Joyful", "Melancholic", "Bright", "Dark", "Gloomy", "Cheerful", "Calm", "Nervous", "Excited", "Anxious", "Serene", "Fierce", "Gentle", "Brave", "Cowardly", "Bold", "Timid", "Strong", "Weak", "Lively", "Dull", "Energetic", "Lethargic", "Optimistic", "Pessimistic", "Confident", "Insecure", "Friendly", "Hostile", "Kind", "Cruel", "Generous", "Selfish", "Humble", "Arrogant", "Polite", "Rude", "Grateful", "Ungrateful", "Patient", "Impatient", "Loyal", "Disloyal", "Trustworthy", "Untrustworthy", "Caring", "Indifferent", "Empathetic", "Apathetic", "Creative", "Unimaginative", "Intelligent", "Foolish", "Wise", "Naive", "Curious", "Indifferent", "Hardworking", "Lazy", "Organized", "Messy", "Reliable", "Unreliable", "Honest", "Deceitful", "Thoughtful", "Thoughtless", "Considerate", "Inconsiderate", "Respectful", "Disrespectful", "Adaptable", "Rigid", "Ambitious", "Unambitious", "Assertive", "Passive", "Attentive", "Distracted", "Charming", "Repellent", "Compassionate", "Heartless", "Determined", "Indecisive", "Disciplined", "Enthusiastic", "Apathetic", "Forgiving", "Vindictive", "Humorous", "Serious", "Imaginative", "Literal", "Loquacious", "Meticulous", "Careless",
}

var nouns = [100]string{
	"Apple", "Banana", "Cherry", "Fig", "Grape", "Jackfruit", "Lemon", "Mango", "Orange", "Quince", "Strawberry", "Ugli fruit", "Watermelon", "Apricot", "Coconut", "Olive", "Avocado", "Cherry", "Dragonfruit", "Mandarin", "Dog", "Elephant", "Horse", "Iguana", "Kangaroo", "Newt", "Penguin", "Rabbit", "Tiger", "Vulture", "Xerus", "Yak", "Zebra", "Bear", "Dolphin", "Eagle", "Giraffe", "Hawk", "Insect", "Jaguar", "Monkey", "Nightingale", "Parrot", "Quail", "Raccoon", "Squirrel", "Tortoise", "Uakari", "Viper", "Walrus", "Yak", "Zebu", "Bat", "Duck", "Emu", "Frog", "Goose", "Hamster", "Ibis", "Jellyfish", "Lion", "Mole", "Ostrich", "Peacock", "Quokka", "Rhinoceros", "Snake", "Toucan", "Urial", "Wolf", "Xenopus", "Yak", "Zebra", "Cheetah", "Echidna", "Fennec", "Guinea pig", "Hedgehog", "Impala", "Jay", "Koala", "Lynx", "Narwhal", "Ocelot", "Panda", "Raven", "Sloth", "Tapir", "Urchin",
}
