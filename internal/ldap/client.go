package ldap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	endpoint   string
	httpClient *http.Client
	authToken  string
}

func GenerateNewClient(endpoint, authToken string) *Client {
	fmt.Println("Generate new client")
	return &Client{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		authToken: authToken,
	}
}

func GenerateNewClientWithAuthToken(endpoint string) *Client {
	fmt.Println("Generating New Client with Auth Token")
	client := &Client{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	client.generateAuthToken("admin", "purple14735#")
	return client
}

type AuthTokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   map[string]interface{} `json:"data,omitempty"`
	Errors []GraphQLError         `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
	Path    []any  `json:"path,omitempty"`
}

func (c *Client) FetchAllUsers(ctx context.Context) (*UsersResponse, error) {
	fmt.Println("Fetching Users")
	query := `
	{
		users {
			id,
			email,
			displayName,
			firstName,
			lastName,
			creationDate,
			uuid,
			groups {
				displayName,
				id
			}
		}
	}`

	req := GraphQLRequest{
		Query: query,
	}

	var response UsersResponse
	err := c.executeQuery(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch users: %w", err)
	}

	fmt.Printf("Users: %s\n", response)

	return &response, nil
}

func (c *Client) generateAuthToken(username, password string) error {
	fmt.Printf("Generating auth token for user %s\n", username)
	authTokenRequest := &AuthTokenRequest{
		Username: username,
		Password: password,
	}
	jsonReq, err := json.Marshal(authTokenRequest)
	if err != nil {
		return fmt.Errorf("Unable to marshal authToken Request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.endpoint+"/auth/simple/login", bytes.NewBuffer(jsonReq))
	if err != nil {
		return fmt.Errorf("Unable to create HTTP request for Auth Token: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("Request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request faled with status: %d", resp.StatusCode)
	}

	var authTokenResponse AuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&authTokenResponse); err != nil {
		return fmt.Errorf("Failed in decoding auth token response data: %w", err)
	}

	c.authToken = authTokenResponse.Token

	return nil
}
func (c *Client) executeQuery(ctx context.Context, req GraphQLRequest, result *UsersResponse) error {
	jsonReq, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("Failed to marshal request: %w", err)
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, jsonReq, "", "\t")
	if err != nil {
		return fmt.Errorf("Unable to prettify JSON request")
	}
	fmt.Println(string(prettyJSON.Bytes()))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/api/graphql", bytes.NewBuffer(jsonReq))
	if err != nil {
		return fmt.Errorf("Failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Origin", c.endpoint)
	if c.authToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	fmt.Printf("HTTP URI: %s\n", httpReq.URL)
	fmt.Printf("HTTP %s\n", httpReq.Header["Authorization"])

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("Request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request faled with status: %d", resp.StatusCode)
	}

	fmt.Printf("Response: %s\n", resp.Body)

	var gqlResp GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return fmt.Errorf("Failed in decoding response data: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return fmt.Errorf("GraphQL Error: %s", gqlResp.Errors[0].Message)
	}

	fmt.Printf("gqlResp: %s\n", gqlResp.Data)

	respData, err := json.Marshal(gqlResp.Data)
	if err != nil {
		return fmt.Errorf("Failed to re-marshal data: %w", err)
	}

	fmt.Println()
	fmt.Printf("Marshalled Data: %s\n", respData)

	if err := json.Unmarshal(respData, result); err != nil {
		return fmt.Errorf("Failed to unmarshal data ino result: %w", err)
	}

	fmt.Println()
	fmt.Printf("Final Data: %s\n", result)

	return nil
}
