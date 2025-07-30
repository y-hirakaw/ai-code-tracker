import Foundation

// Simple vocabulary learning app
struct VocabularyCard {
    let word: String
    let translation: String
}

let cards = [
    VocabularyCard(word: "Hello", translation: "こんにちは"),
    VocabularyCard(word: "Thank you", translation: "ありがとう")
    VocabularyCard(word: "Fine?", translation: "元気？")
    VocabularyCard(word: "Good bye", translation: "さようなら")
]

print("Learning \(cards.count) words")