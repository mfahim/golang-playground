package main

import (
    "encoding/csv"
    "fmt"
    "os"
    "strings"
)

// Problem represents a single quiz question and its answer
type Problem struct {
    Question string
    Answer   string
}

func main() {
    // Open the CSV file
    file, err := os.Open("problems.csv")
    if err != nil {
        fmt.Printf("Failed to open CSV file: %v\n", err)
        os.Exit(1)
    }
    defer file.Close()

    // Create a new CSV reader
    reader := csv.NewReader(file)

    // Read all records from CSV
    records, err := reader.ReadAll()
    if err != nil {
        fmt.Printf("Failed to read CSV: %v\n", err)
        os.Exit(1)
    }

    // Convert records to problems
    problems := parseProblems(records)

    // Initialize score counter
    score := 0
    totalQuestions := len(problems)

    // Run the quiz
    fmt.Println("Welcome to the Quiz!")
    fmt.Println("Answer each question and press Enter.")
    fmt.Println("-----------------------------------")

    for i, prob := range problems {
        // Display question
        fmt.Printf("Question #%d: %s\n", i+1, prob.Question)
        fmt.Print("Your answer: ")

        // Get user's answer
        var userAnswer string
        fmt.Scanln(&userAnswer)

        // Check answer (case-insensitive)
        if strings.EqualFold(strings.TrimSpace(userAnswer), strings.TrimSpace(prob.Answer)) {
            score++
        }
    }

    // Display final results
    fmt.Println("\n-----------------------------------")
    fmt.Printf("Quiz completed!\n")
    fmt.Printf("You scored %d out of %d questions correctly\n", score, totalQuestions)
    fmt.Printf("Percentage: %.2f%%\n", float64(score)/float64(totalQuestions)*100)
}

// parseProblems converts CSV records to Problem structs
func parseProblems(records [][]string) []Problem {
    problems := make([]Problem, 0, len(records))
    for _, record := range records {
        if len(record) != 2 {
            fmt.Printf("Warning: skipping invalid record: %v\n", record)
            continue
        }
        problems = append(problems, Problem{
            Question: record[0],
            Answer:   record[1],
        })
    }
    return problems
}