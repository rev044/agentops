package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

// Message represents an inter-agent message.
type Message struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Body      string    `json:"body"`
	Timestamp time.Time `json:"timestamp"`
	Read      bool      `json:"read"`
	Type      string    `json:"type"` // progress, completion, blocker, farm_complete
}

var (
	inboxSince    string
	inboxFrom     string
	inboxUnread   bool
	inboxMarkRead bool
	mailTo        string
	mailBody      string
	mailType      string
)

var inboxCmd = &cobra.Command{
	Use:   "inbox",
	Short: "Check messages from agents",
	Long: `View messages from the Agent Farm.

Messages include:
  - Progress summaries from witness
  - Completion notifications from agents
  - Blocker escalations
  - Farm complete signal

Examples:
  ao inbox
  ao inbox --since 5m
  ao inbox --from witness
  ao inbox --unread`,
	RunE: runInbox,
}

var mailCmd = &cobra.Command{
	Use:   "mail",
	Short: "Send and receive agent messages",
	Long: `Inter-agent messaging for the Agent Farm.

Commands:
  send    Send a message
  inbox   View received messages (alias for ao inbox)

Examples:
  ao mail send --to mayor --body "Issue complete"
  ao mail send --to mayor --body "FARM COMPLETE" --type farm_complete`,
}

var mailSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a message",
	Long: `Send a message to another agent or the mayor.

Examples:
  ao mail send --to mayor --body "Completed issue gt-123"
  ao mail send --to witness --body "Agent 1 stuck"
  ao mail send --to mayor --body "FARM COMPLETE" --type farm_complete`,
	RunE: runMailSend,
}

func init() {
	rootCmd.AddCommand(inboxCmd)
	rootCmd.AddCommand(mailCmd)

	mailCmd.AddCommand(mailSendCmd)

	// Inbox flags
	inboxCmd.Flags().StringVar(&inboxSince, "since", "", "Show messages from last duration (e.g., 5m, 1h)")
	inboxCmd.Flags().StringVar(&inboxFrom, "from", "", "Filter by sender")
	inboxCmd.Flags().BoolVar(&inboxUnread, "unread", false, "Show only unread messages")
	inboxCmd.Flags().BoolVar(&inboxMarkRead, "mark-read", false, "Mark displayed messages as read")

	// Mail send flags
	mailSendCmd.Flags().StringVar(&mailTo, "to", "", "Recipient (mayor, witness, agent-N)")
	mailSendCmd.Flags().StringVar(&mailBody, "body", "", "Message body")
	mailSendCmd.Flags().StringVar(&mailType, "type", "progress", "Message type (progress, completion, blocker, farm_complete)")

	_ = mailSendCmd.MarkFlagRequired("to")
	_ = mailSendCmd.MarkFlagRequired("body")
}

func runInbox(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// Load messages
	messages, err := loadMessages(cwd)
	if err != nil {
		// If no messages file, show empty
		if os.IsNotExist(err) {
			fmt.Println("No messages")
			return nil
		}
		return fmt.Errorf("load messages: %w", err)
	}

	// Filter messages
	filtered := filterMessages(messages, inboxSince, inboxFrom, inboxUnread)

	if len(filtered) == 0 {
		fmt.Println("No messages")
		return nil
	}

	// Sort by timestamp descending (newest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// Output based on format
	switch GetOutput() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(filtered)

	default:
		// Table format
		fmt.Println()
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TIME\tFROM\tTYPE\tMESSAGE")
		fmt.Fprintln(w, "----\t----\t----\t-------")

		for _, msg := range filtered {
			age := formatAge(msg.Timestamp)
			body := truncateMessage(msg.Body, 60)
			unreadMark := ""
			if !msg.Read {
				unreadMark = "*"
			}
			fmt.Fprintf(w, "%s%s\t%s\t%s\t%s\n", unreadMark, age, msg.From, msg.Type, body)
		}

		w.Flush()
		fmt.Printf("\n%d message(s)\n", len(filtered))
	}

	// Mark as read if requested
	if inboxMarkRead {
		if err := markMessagesRead(cwd, filtered); err != nil {
			VerbosePrintf("Warning: failed to mark messages as read: %v\n", err)
		}
	}

	return nil
}

func runMailSend(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// Determine sender identity
	from := os.Getenv("AO_AGENT_NAME")
	if from == "" {
		from = "unknown"
	}

	// Create message
	msg := Message{
		ID:        generateMessageID(),
		From:      from,
		To:        mailTo,
		Body:      mailBody,
		Timestamp: time.Now(),
		Read:      false,
		Type:      mailType,
	}

	if GetDryRun() {
		fmt.Printf("[dry-run] Would send message:\n")
		fmt.Printf("  From: %s\n", msg.From)
		fmt.Printf("  To: %s\n", msg.To)
		fmt.Printf("  Type: %s\n", msg.Type)
		fmt.Printf("  Body: %s\n", msg.Body)
		return nil
	}

	// Append to messages file
	if err := appendMessage(cwd, &msg); err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	fmt.Printf("Message sent to %s\n", mailTo)
	VerbosePrintf("ID: %s\n", msg.ID)

	return nil
}

// Helper functions

func loadMessages(cwd string) ([]Message, error) {
	messagesPath := filepath.Join(cwd, ".agents", "mail", "messages.jsonl")
	file, err := os.Open(messagesPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var messages []Message
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var msg Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return messages, scanner.Err()
}

func filterMessages(messages []Message, since, from string, unreadOnly bool) []Message {
	var filtered []Message

	// Parse since duration
	var sinceTime time.Time
	if since != "" {
		duration, err := time.ParseDuration(since)
		if err == nil {
			sinceTime = time.Now().Add(-duration)
		}
	}

	for _, msg := range messages {
		// Filter by time
		if !sinceTime.IsZero() && msg.Timestamp.Before(sinceTime) {
			continue
		}

		// Filter by sender
		if from != "" && msg.From != from {
			continue
		}

		// Filter by unread
		if unreadOnly && msg.Read {
			continue
		}

		// Default: show messages to "mayor" or "all"
		if msg.To != "mayor" && msg.To != "all" && msg.To != "" {
			continue
		}

		filtered = append(filtered, msg)
	}

	return filtered
}

func appendMessage(cwd string, msg *Message) error {
	mailDir := filepath.Join(cwd, ".agents", "mail")
	if err := os.MkdirAll(mailDir, 0700); err != nil {
		return err
	}

	messagesPath := filepath.Join(mailDir, "messages.jsonl")

	// Append to file
	file, err := os.OpenFile(messagesPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if _, err := file.WriteString(string(data) + "\n"); err != nil {
		return err
	}

	return nil
}

func markMessagesRead(cwd string, messages []Message) error {
	// Load all messages
	allMessages, err := loadMessages(cwd)
	if err != nil {
		return err
	}

	// Create a set of IDs to mark
	toMark := make(map[string]bool)
	for _, msg := range messages {
		toMark[msg.ID] = true
	}

	// Update messages
	for i := range allMessages {
		if toMark[allMessages[i].ID] {
			allMessages[i].Read = true
		}
	}

	// Rewrite file
	messagesPath := filepath.Join(cwd, ".agents", "mail", "messages.jsonl")
	file, err := os.Create(messagesPath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, msg := range allMessages {
		data, err := json.Marshal(msg)
		if err != nil {
			continue
		}
		file.WriteString(string(data) + "\n")
	}

	return nil
}

func generateMessageID() string {
	return fmt.Sprintf("msg-%d", time.Now().UnixNano())
}

func formatAge(t time.Time) string {
	age := time.Since(t)

	if age < time.Minute {
		return fmt.Sprintf("%ds ago", int(age.Seconds()))
	}
	if age < time.Hour {
		return fmt.Sprintf("%dm ago", int(age.Minutes()))
	}
	if age < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(age.Hours()))
	}
	return t.Format("Jan 2")
}

func truncateMessage(s string, max int) string {
	// Replace newlines with spaces
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)

	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
