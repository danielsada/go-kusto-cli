package input

import (
	"regexp"
	"strings"
)

// clusterRe matches cluster('...') or cluster("...") in KQL.
var clusterRe = regexp.MustCompile(`cluster\(\s*['"]([^'"]+)['"]\s*\)`)

// databaseRe matches database('...') or database("...") in KQL.
var databaseRe = regexp.MustCompile(`database\(\s*['"]([^'"]+)['"]\s*\)`)

// InferredParams holds cluster and database values extracted from KQL text.
type InferredParams struct {
	Cluster  string
	Database string
}

// InferParams scans one or more KQL queries for cluster('...') and database('...')
// expressions and returns the first match found for each.
func InferParams(queries []string) InferredParams {
	var params InferredParams
	for _, q := range queries {
		if params.Cluster == "" {
			if m := clusterRe.FindStringSubmatch(q); len(m) > 1 {
				params.Cluster = normalizeClusterURL(m[1])
			}
		}
		if params.Database == "" {
			if m := databaseRe.FindStringSubmatch(q); len(m) > 1 {
				params.Database = m[1]
			}
		}
		if params.Cluster != "" && params.Database != "" {
			break
		}
	}
	return params
}

// normalizeClusterURL ensures the cluster value is a full URL.
// KQL cluster() can contain just a name like "mycluster" or a full URL.
func normalizeClusterURL(raw string) string {
	if strings.HasPrefix(raw, "https://") || strings.HasPrefix(raw, "http://") {
		return raw
	}
	// Bare cluster name — assume it's an Azure Data Explorer cluster.
	// e.g. "mycluster" → "https://mycluster.kusto.windows.net"
	return "https://" + raw + ".kusto.windows.net"
}
