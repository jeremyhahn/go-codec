package main

// fleetHosts holds the company fleet's hostnames as salted SHA-256 digests
// (HostHash), never as names: a gate that listed them would itself publish
// them. The source of the list is the handbook's server and network
// inventories (infra/servers/README.md, infra/network/docs/01-inventory.md).
//
// A digest matches a whole token of letters, digits and inner hyphens, and a
// "family" digest also matches that stem followed by -<number>. Names too
// common to match as bare words belong in contextHosts instead.
//
// To add a machine: go run ./internal/cmd/releasability -hash <name>
var fleetHosts = map[string]struct{}{
	// Exact hostnames.
	"e3e6f24e31cf9d7de315e7f43141628c3315462849b54382e85ddaac6ea03e28": {},
	"2bbe03177cd149e0422e23cd892179cb4765f0db919f2411f084d4f7c5578ac5": {},
	"db97e90ff3377a8fecac9b7da5ab221f7b527d04c54f11224b34cb57add9bfec": {},
	"73137196a67ed9096e410767a04456b8ed049317a05ffeeabc38b36c594438f3": {},
	"9a7c73f46103248e2859cb01cfb1e7df37fc296f6280bc0c0252212cc6ba2b77": {},
	"0aba308b39fe1e202adbd02debd2b7d6f6ce70a6d7279d7ac81a7f0319cfe7b8": {},
	"80227528385dbd0f971b6ed901d427f35e33f76b1bac6ffcd3c589b21a5c0248": {},
	"fef7a99bfee69f7c6f4cb852bd3e92bc5ddbe352b15938081fc9da867bf09185": {},
	"0fdab79d08f89648e4463fba48cfcfb5b2911b195b4bea2775d93cf059a81428": {},
	"898d8147a0f29acd4f5256446cd70347b283439bba131e34c9c4506e78a45007": {},
	// Family stems: the hypervisors, workstations and CI guests, routers,
	// switches, KVMs and UPSes, each numbered -NN.
	"b1e727fb6503dd16335fb53c6aa488d8d17512ebceadafc7855c10649b3f5b63": {},
	"28baa64607f6566e8a15fb813f7fd391d0906e62db6cae42a971bdea6938b3c1": {},
	"073446559b183a64aa55ffe06fdeb38d32f64d9ac775aadde3ef2cca547f5feb": {},
	"690d610d452f8493b12da661f9d18714aa584cf25a59f3b92d9fba412192030b": {},
	"30b0862906aa1588838e1266603d28b335dc9e30d048890edeb3e24359278346": {},
	"09b31cb7647462949b29e74864a22736ac847286d1a3fb0b25a01f4ffc09fb04": {},
	"3e8cb9af833d46b172d017a72d186e92015892647e8f187331420201fd653b56": {},
	"d6a2cab8dc27f6b8593fa277ae901eb05cd64b908ce55de1db2e102dee9cbb45": {},
	"520305c59429ec5f88249818b9c27cc745912bb4f8d24f93b2aac15879cf9dab": {},
	"a56e9976b9e2c84ca3aac30aa64bf7a71b6e848399ee0adb26acf07c4a1c0d60": {},
}

// contextHosts holds, as HostHash digests, the fleet hostnames that are also
// ordinary words or product names (the fileserver, the Gitea guest, the
// secondary compute host). A bare-word match would fail on prose and code that
// merely mention the word, so these match only where the text uses the name as
// a host: a URL authority, a UNC path, user@host, an ssh/scp/rsync/ping-style
// command argument, host:path or host:port, an ssh-config Host line, a
// host/hostname/server setting, or name.lan/.local. See hostContextPatterns
// and hostCommand.
var contextHosts = map[string]struct{}{
	"b922f9f95f307d029177bb8a0aed0553ba187f4df1ac5c00910f4205f59f2b3b": {},
	"9bc90d09ed7bec3c56ea98b51ae53808c747f393365e1d11dfea5539a9dacb1e": {},
	"ca4da804fcf103937810a38ff2f0d02e92e614bb4ecc4d131a9d68d5359fb171": {},
}
