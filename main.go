package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	virtual_fido "github.com/bulwarkid/virtual-fido"
	"github.com/bulwarkid/virtual-fido/fido_client"
	"github.com/spf13/cobra"
)

var vaultFilename string
var vaultPassphrase string
var identityID string
var verbose bool

func checkErr(err error, message string) {
	if err != nil {
		panic(fmt.Sprintf("Error: %s - %s", err, message))
	}
}

func listIdentities(cmd *cobra.Command, args []string) {
	client := createClient()
	fmt.Printf("------- Identities in file '%s' -------\n", vaultFilename)
	sources := client.Identities()
	for _, source := range sources {
		fmt.Printf("(%s): '%s' for website '%s'\n", hex.EncodeToString(source.ID[:4]), source.User.Name, source.RelyingParty.Name)
	}
}

func deleteIdentity(cmd *cobra.Command, args []string) {
	client := createClient()
	ids := client.Identities()
	targetIDs := make([]*fido_client.CredentialSource, 0)
	for _, id := range ids {
		hexString := hex.EncodeToString(id.ID)
		if strings.HasPrefix(hexString, identityID) {
			targetIDs = append(targetIDs, &id)
		}
	}
	if len(targetIDs) > 1 {
		fmt.Printf("Multiple identities with prefix (%s):\n", identityID)
		for _, id := range targetIDs {
			fmt.Printf("- (%s)\n", hex.EncodeToString(id.ID))
		}
	} else if len(targetIDs) == 1 {
		fmt.Printf("Deleting identity (%s)\n...", hex.EncodeToString(targetIDs[0].ID))
		if client.DeleteIdentity(targetIDs[0].ID) {
			fmt.Printf("Done.\n")
		} else {
			fmt.Printf("Could not find (%s).\n", hex.EncodeToString(targetIDs[0].ID))
		}
	} else {
		fmt.Printf("No identity found with prefix (%s)\n", identityID)
	}
}

func enablePIN(cmd *cobra.Command, args []string) {
	client := createClient()
	if client.PINHash() == nil {
		cmd.PrintErr("Set a PIN first with `pin set --pin <4+ digits>`\n")
		return
	}
	cmd.Println("PIN enabled")
}

func disablePIN(cmd *cobra.Command, args []string) {
	client := createClient()
	client.SetPINHash(nil)
	cmd.Println("PIN disabled")
}

var newPIN int

func setPIN(cmd *cobra.Command, args []string) {
	if newPIN < 0 {
		cmd.PrintErr("Invalid PIN: PIN must be positive")
		return
	}
	newPINString := strconv.Itoa(newPIN)
	if len(newPINString) < 4 {
		cmd.PrintErr("Invalid PIN: PIN must be 4 digits")
		return
	}
	client := createClient()
	pinHash := sha256.Sum256([]byte(newPINString))
	client.SetPINHash(pinHash[:16])
	cmd.Println("PIN set")
}

func start(cmd *cobra.Command, args []string) {
	log.Printf("Starting FakeFido")
	if verbose {
		log.Printf("Vault: %s", vaultFilename)
	}
	client := createClient()
	runServer(client)
	log.Printf("FakeFido stopped")
}

func createClient() *fido_client.DefaultFIDOClient {
	// ALL OF THIS IS INSECURE, FOR TESTING PURPOSES ONLY
	caPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	checkErr(err, "Could not generate attestation CA private key")
	certificateAuthority := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"FakeFido"}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	certificateBytes, err := x509.CreateCertificate(rand.Reader, certificateAuthority, certificateAuthority, &caPrivateKey.PublicKey, caPrivateKey)
	checkErr(err, "Could not create attestation CA certificate")
	encryptionKey := sha256.Sum256([]byte("test"))

	virtual_fido.SetLogOutput(os.Stdout)
	support := ClientSupport{vaultFilename: vaultFilename, vaultPassphrase: vaultPassphrase}
	return fido_client.NewDefaultClient(certificateBytes, caPrivateKey, encryptionKey, &support, &support)
}

var rootCmd = &cobra.Command{
	Use:   "demo",
	Short: "Run Virtual FIDO demo",
	Long:  `demo attaches a virtual FIDO2 device for logging in with WebAuthN`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&vaultFilename, "vault", "", "vault.json", "Identity vault filename")
	rootCmd.PersistentFlags().StringVarP(&vaultPassphrase, "passphrase", "", "passphrase", "Identity vault passphrase")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.MarkFlagRequired("vault")
	rootCmd.MarkFlagRequired("passphrase")
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	start := &cobra.Command{
		Use:   "start",
		Short: "Attach virtual FIDO device",
		Run:   start,
	}
	rootCmd.AddCommand(start)

	list := &cobra.Command{
		Use:   "list",
		Short: "List identities in vault",
		Run:   listIdentities,
	}
	rootCmd.AddCommand(list)

	delete := &cobra.Command{
		Use:   "delete",
		Short: "Delete identity in vault",
		Run:   deleteIdentity,
	}
	delete.Flags().StringVar(&identityID, "identity", "", "Identity hash to delete")
	delete.MarkFlagRequired("identity")
	rootCmd.AddCommand(delete)

	pinCommand := &cobra.Command{
		Use:   "pin",
		Short: "Modify PIN Behavior",
	}
	enablePINCommand := &cobra.Command{
		Use:   "enable",
		Short: "Enables PIN protection",
		Run:   enablePIN,
	}
	pinCommand.AddCommand(enablePINCommand)
	disablePINCommand := &cobra.Command{
		Use:   "disable",
		Short: "Disables PIN protection",
		Run:   disablePIN,
	}
	pinCommand.AddCommand(disablePINCommand)
	setPINCommand := &cobra.Command{
		Use:   "set",
		Short: "Sets the PIN",
		Run:   setPIN,
	}
	setPINCommand.Flags().IntVar(&newPIN, "pin", -1, "New PIN")
	setPINCommand.MarkFlagRequired("pin")
	pinCommand.AddCommand(setPINCommand)
	rootCmd.AddCommand(pinCommand)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
