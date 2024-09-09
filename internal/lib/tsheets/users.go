package tsheets

type UsersQueryParams struct {
	IDs              string `url:"ids,omitempty"`               // Comma-separated list of user IDs to filter by
	NotIDs           string `url:"not_ids,omitempty"`           // Comma-separated list of user IDs to exclude
	EmployeeNumbers  string `url:"employee_numbers,omitempty"`  // Comma-separated list of employee numbers to filter by
	Usernames        string `url:"usernames,omitempty"`         // Comma-separated list of usernames to filter by
	GroupIDs         string `url:"group_ids,omitempty"`         // Comma-separated list of group IDs to filter by
	NotGroupIDs      string `url:"not_group_ids,omitempty"`     // Comma-separated list of group IDs to exclude
	PayrollIDs       string `url:"payroll_ids,omitempty"`       // Comma-separated string of payroll IDs to filter by
	Active           string `url:"active,omitempty"`            // 'yes', 'no', or 'both'. Default is 'yes'
	FirstName        string `url:"first_name,omitempty"`        // Wildcard search for first names
	LastName         string `url:"last_name,omitempty"`         // Wildcard search for last names
	ModifiedBefore   string `url:"modified_before,omitempty"`   // Return only users modified before this date (ISO 8601 format)
	ModifiedSince    string `url:"modified_since,omitempty"`    // Return only users modified since this date (ISO 8601 format)
	SupplementalData string `url:"supplemental_data,omitempty"` // 'yes' or 'no', indicates if supplemental data should be included (default: 'yes')
	Limit            int    `url:"limit,omitempty"`             // Number of results to retrieve per request (1-200, default 200)
	Page             int    `url:"page,omitempty"`              // Represents the page of results to retrieve (default 1)
}

func (p UsersQueryParams) getUsers() (struct{}, error)
