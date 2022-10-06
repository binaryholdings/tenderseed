package main

type Chains struct {
	Chains []string `json:"chains"`
}

type Chain struct {
	Schema       string `json:"$schema,omitempty"`
	ChainName    string `json:"chain_name,omitempty"`
	Status       string `json:"status,omitempty"`
	NetworkType  string `json:"network_type,omitempty"`
	PrettyName   string `json:"pretty_name,omitempty"`
	ChainID      string `json:"chain_id,omitempty"`
	Bech32Prefix string `json:"bech32_prefix,omitempty"`
	DaemonName   string `json:"daemon_name,omitempty"`
	NodeHome     string `json:"node_home,omitempty"`
	Genesis      struct {
		GenesisURL string `json:"genesis_url,omitempty"`
	} `json:"genesis,omitempty"`
	Slip44   int `json:"slip44,omitempty"`
	Codebase struct {
		GitRepo            string   `json:"git_repo,omitempty"`
		RecommendedVersion string   `json:"recommended_version,omitempty"`
		CompatibleVersions []string `json:"compatible_versions,omitempty"`
		Binaries           struct {
			LinuxAmd64 string `json:"linux/amd64,omitempty"`
		} `json:"binaries,omitempty"`
	} `json:"codebase,omitempty"`
	Peers struct {
		Seeds []struct {
			ID       string `json:"id,omitempty"`
			Address  string `json:"address,omitempty"`
			Provider string `json:"provider,omitempty"`
		} `json:"seeds,omitempty"`
		PersistentPeers []struct {
			ID      string `json:"id,omitempty"`
			Address string `json:"address,omitempty"`
		} `json:"persistent_peers,omitempty"`
	} `json:"peers,omitempty"`
	Apis struct {
		RPC []struct {
			Address  string `json:"address,omitempty"`
			Provider string `json:"provider,omitempty"`
		} `json:"rpc,omitempty"`
		Rest []struct {
			Address string `json:"address,omitempty"`
		} `json:"rest,omitempty"`
	} `json:"apis,omitempty"`
}
