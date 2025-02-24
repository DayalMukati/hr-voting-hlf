package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a voting system
type SmartContract struct {
	contractapi.Contract
}

// Voter represents a registered voter
type Voter struct {
	ID         string `json:"ID"`
	Name       string `json:"Name"`
	Eligibility bool   `json:"Eligibility"`
	HasVoted   bool   `json:"HasVoted"`
}

// Election represents an election with candidates and votes
type Election struct {
	ID        string            `json:"ID"`
	Title     string            `json:"Title"`
	Candidates []string          `json:"Candidates"`
	Votes     map[string]int    `json:"Votes"`
	StartTime time.Time         `json:"StartTime"`
	EndTime   time.Time         `json:"EndTime"`
}

// RegisterVoter enrolls a new voter
func (s *SmartContract) RegisterVoter(ctx contractapi.TransactionContextInterface, voterID string, name string) error {
	voter := Voter{
		ID:         voterID,
		Name:       name,
		Eligibility: true,
		HasVoted:   false,
	}

	voterJSON, err := json.Marshal(voter)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(voterID, voterJSON)
}

// CreateElection initiates a new election
func (s *SmartContract) CreateElection(ctx contractapi.TransactionContextInterface, electionID string, title string, candidates []string, startTime string, endTime string) error {
	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return fmt.Errorf("invalid start time format: %v", err)
	}

	end, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return fmt.Errorf("invalid end time format: %v", err)
	}

	if end.Before(start) {
		return fmt.Errorf("end time must be after start time")
	}

	votes := make(map[string]int)
	for _, candidate := range candidates {
		votes[candidate] = 0
	}

	election := Election{
		ID:        electionID,
		Title:     title,
		Candidates: candidates,
		Votes:     votes,
		StartTime: start,
		EndTime:   end,
	}

	electionJSON, err := json.Marshal(election)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(electionID, electionJSON)
}

// CastVote allows a voter to cast a vote for a candidate in an election
func (s *SmartContract) CastVote(ctx contractapi.TransactionContextInterface, voterID string, electionID string, candidate string) error {
	voterJSON, err := ctx.GetStub().GetState(voterID)
	if err != nil {
		return fmt.Errorf("failed to read voter: %v", err)
	}
	if voterJSON == nil {
		return fmt.Errorf("voter %s does not exist", voterID)
	}

	var voter Voter
	err = json.Unmarshal(voterJSON, &voter)
	if err != nil {
		return err
	}

	if !voter.Eligibility {
		return fmt.Errorf("voter %s is not eligible to vote", voterID)
	}
	if voter.HasVoted {
		return fmt.Errorf("voter %s has already voted", voterID)
	}

	electionJSON, err := ctx.GetStub().GetState(electionID)
	if err != nil {
		return fmt.Errorf("failed to read election: %v", err)
	}
	if electionJSON == nil {
		return fmt.Errorf("election %s does not exist", electionID)
	}

	var election Election
	err = json.Unmarshal(electionJSON, &election)
	if err != nil {
		return err
	}

	currentTime := time.Now()
	if currentTime.Before(election.StartTime) || currentTime.After(election.EndTime) {
		return fmt.Errorf("election %s is not active", electionID)
	}

	if _, exists := election.Votes[candidate]; !exists {
		return fmt.Errorf("candidate %s is not in the election %s", candidate, electionID)
	}

	election.Votes[candidate]++
	voter.HasVoted = true

	electionJSON, err = json.Marshal(election)
	if err != nil {
		return err
	}

	voterJSON, err = json.Marshal(voter)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(electionID, electionJSON)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(voterID, voterJSON)
}

// TallyVotes counts the votes for each candidate in an election
func (s *SmartContract) TallyVotes(ctx contractapi.TransactionContextInterface, electionID string) (map[string]int, error) {
	electionJSON, err := ctx.GetStub().GetState(electionID)
	if err != nil {
		return nil, fmt.Errorf("failed to read election: %v", err)
	}
	if electionJSON == nil {
		return nil, fmt.Errorf("election %s does not exist", electionID)
	}

	var election Election
	err = json.Unmarshal(electionJSON, &election)
	if err != nil {
		return nil, err
	}

	currentTime := time.Now()
	if currentTime.Before(election.EndTime) {
		return nil, fmt.Errorf("election %s is still ongoing", electionID)
	}

	return election.Votes, nil
}

// GetElectionResults retrieves the final results of an election
func (s *SmartContract) GetElectionResults(ctx contractapi.TransactionContextInterface, electionID string) (map[string]int, error) {
	return s.TallyVotes(ctx, electionID)
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Error creating voting chaincode: %s", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting voting chaincode: %s", err)
	}
}
