package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	authparams "github.com/treeverse/lakefs/auth/params"
	"github.com/treeverse/lakefs/block/factory"
	blockparams "github.com/treeverse/lakefs/block/params"
	catalogparams "github.com/treeverse/lakefs/catalog/mvcc/params"
	dbparams "github.com/treeverse/lakefs/db/params"
	"github.com/treeverse/lakefs/logging"
	pyramidparams "github.com/treeverse/lakefs/pyramid/params"
)

const (
	DefaultBlockStoreType                    = "local"
	DefaultBlockStoreLocalPath               = "~/lakefs/data"
	DefaultBlockStoreS3Region                = "us-east-1"
	DefaultBlockStoreS3StreamingChunkSize    = 2 << 19         // 1MiB by default per chunk
	DefaultBlockStoreS3StreamingChunkTimeout = time.Second * 1 // or 1 seconds, whatever comes first

	DefaultCommittedLocalCacheBytes    = 1 * 1024 * 1024 * 1024
	DefaultCommittedLocalCacheDir      = "~/lakefs/local_tier"
	DefaultCommittedBlockStoragePrefix = "_lakefs"

	DefaultBlockStoreGSS3Endpoint = "https://storage.googleapis.com"

	DefaultAuthCacheEnabled = true
	DefaultAuthCacheSize    = 1024
	DefaultAuthCacheTTL     = 20 * time.Second
	DefaultAuthCacheJitter  = 3 * time.Second

	DefaultListenAddr          = "0.0.0.0:8000"
	DefaultS3GatewayDomainName = "s3.local.lakefs.io"
	DefaultS3GatewayRegion     = "us-east-1"
	DefaultS3MaxRetries        = 5

	DefaultStatsEnabled       = true
	DefaultStatsAddr          = "https://stats.treeverse.io"
	DefaultStatsFlushInterval = time.Second * 30

	MetaStoreType          = "metastore.type"
	MetaStoreHiveURI       = "metastore.hive.uri"
	MetastoreGlueCatalogID = "metastore.glue.catalog_id"
)

var (
	ErrMissingSecretKey = errors.New("auth.encrypt.secret_key cannot be empty")
)

type LogrusAWSAdapter struct {
	logger *log.Entry
}

func (l *LogrusAWSAdapter) Log(vars ...interface{}) {
	l.logger.Debug(vars...)
}

type Config struct{}

func NewConfig() *Config {
	setDefaults()
	setupLogger()
	return &Config{}
}

func setDefaults() {
	viper.SetDefault("listen_address", DefaultListenAddr)

	viper.SetDefault("logging.format", DefaultLoggingFormat)
	viper.SetDefault("logging.level", DefaultLoggingLevel)
	viper.SetDefault("logging.output", DefaultLoggingOutput)

	viper.SetDefault("auth.cache.enabled", DefaultAuthCacheEnabled)
	viper.SetDefault("auth.cache.size", DefaultAuthCacheSize)
	viper.SetDefault("auth.cache.ttl", DefaultAuthCacheTTL)
	viper.SetDefault("auth.cache.jitter", DefaultAuthCacheJitter)

	viper.SetDefault("blockstore.type", DefaultBlockStoreType)
	viper.SetDefault("blockstore.local.path", DefaultBlockStoreLocalPath)
	viper.SetDefault("blockstore.s3.region", DefaultBlockStoreS3Region)
	viper.SetDefault("blockstore.s3.streaming_chunk_size", DefaultBlockStoreS3StreamingChunkSize)
	viper.SetDefault("blockstore.s3.streaming_chunk_timeout", DefaultBlockStoreS3StreamingChunkTimeout)
	viper.SetDefault("blockstore.s3.max_retries", DefaultS3MaxRetries)

	viper.SetDefault("committed.local_cache.size_bytes", DefaultCommittedLocalCacheBytes)
	viper.SetDefault("committed.local_cache.dir", DefaultCommittedLocalCacheDir)
	viper.SetDefault("committed.block_storage_prefix", DefaultCommittedBlockStoragePrefix)

	viper.SetDefault("gateways.s3.domain_name", DefaultS3GatewayDomainName)
	viper.SetDefault("gateways.s3.region", DefaultS3GatewayRegion)

	viper.SetDefault("blockstore.gs.s3_endpoint", DefaultBlockStoreGSS3Endpoint)

	viper.SetDefault("stats.enabled", DefaultStatsEnabled)
	viper.SetDefault("stats.address", DefaultStatsAddr)
	viper.SetDefault("stats.flush_interval", DefaultStatsFlushInterval)
}

func (c *Config) GetDatabaseParams() dbparams.Database {
	return dbparams.Database{
		ConnectionString:      viper.GetString("database.connection_string"),
		MaxOpenConnections:    viper.GetInt32("database.max_open_connections"),
		MaxIdleConnections:    viper.GetInt32("database.max_idle_connections"),
		ConnectionMaxLifetime: viper.GetDuration("database.connection_max_lifetime"),
	}
}

func (c *Config) GetCatalogerType() string {
	return viper.GetString("cataloger.type")
}

func (c *Config) GetMvccCatalogerCatalogParams() catalogparams.Catalog {
	return catalogparams.Catalog{
		BatchRead: catalogparams.BatchRead{
			EntryMaxWait:  viper.GetDuration("cataloger.batch_read.read_entry_max_wait"),
			ScanTimeout:   viper.GetDuration("cataloger.batch_read.scan_timeout"),
			Delay:         viper.GetDuration("cataloger.batch_read.batch_delay"),
			EntriesAtOnce: viper.GetInt("cataloger.batch_read.entries_read_at_once"),
			Readers:       viper.GetInt("cataloger.batch_read.readers"),
		},
		BatchWrite: catalogparams.BatchWrite{
			EntriesInsertSize: viper.GetInt("cataloger.batch_write.insert_size"),
		},
		Cache: catalogparams.Cache{
			Enabled: viper.GetBool("cataloger.cache.enabled"),
			Size:    viper.GetInt("cataloger.cache.size"),
			Expiry:  viper.GetDuration("cataloger.cache.expiry"),
			Jitter:  viper.GetDuration("cataloger.cache.jitter"),
		},
	}
}

type AwsS3RetentionConfig struct {
	RoleArn           string
	ManifestBaseURL   *url.URL
	ReportS3PrefixURL *string
}

func (c *Config) GetAwsS3RetentionConfig() AwsS3RetentionConfig {
	var errs []string
	roleArn := viper.GetString("blockstore.s3.retention.role_arn")
	if roleArn == "" {
		errs = append(errs, "blockstore.s3.retention.role_arn")
	}

	manifestBaseURL, err := url.ParseRequestURI(viper.GetString("blockstore.s3.retention.manifest_base_url"))
	if err != nil {
		errs = append(errs, fmt.Sprintf("blockstore.s3.retention.manifest_base_url: %s", err))
	}
	if len(errs) > 0 {
		panic(fmt.Sprintf("need %s to handle retention on S3", strings.Join(errs, ", ")))
	}
	var reportS3PrefixURL *string
	prefixURL := viper.GetString("blockstore.s3.retention.report_s3_prefix_url")
	if prefixURL != "" {
		reportS3PrefixURL = &prefixURL
	}
	return AwsS3RetentionConfig{
		RoleArn:           roleArn,
		ManifestBaseURL:   manifestBaseURL,
		ReportS3PrefixURL: reportS3PrefixURL,
	}
}

func (c *Config) GetAwsConfig() *aws.Config {
	cfg := &aws.Config{
		Region: aws.String(viper.GetString("blockstore.s3.region")),
		Logger: &LogrusAWSAdapter{log.WithField("sdk", "aws")},
	}
	level := strings.ToLower(logging.Level())
	if level == "trace" {
		cfg.LogLevel = aws.LogLevel(aws.LogDebugWithRequestRetries | aws.LogDebugWithRequestErrors)
	}
	if viper.IsSet("blockstore.s3.profile") || viper.IsSet("blockstore.s3.credentials_file") {
		cfg.Credentials = credentials.NewSharedCredentials(
			viper.GetString("blockstore.s3.credentials_file"),
			viper.GetString("blockstore.s3.profile"))
	}
	if viper.IsSet("blockstore.s3.credentials.access_key_id") {
		cfg.Credentials = credentials.NewStaticCredentials(
			viper.GetString("blockstore.s3.credentials.access_key_id"),
			viper.GetString("blockstore.s3.credentials.access_secret_key"),
			viper.GetString("blockstore.s3.credentials.session_token"))
	}

	s3Endpoint := viper.GetString("blockstore.s3.endpoint")
	if len(s3Endpoint) > 0 {
		cfg = cfg.WithEndpoint(s3Endpoint)
	}
	s3ForcePathStyle := viper.GetBool("blockstore.s3.force_path_style")
	if s3ForcePathStyle {
		cfg = cfg.WithS3ForcePathStyle(true)
	}
	cfg.WithMaxRetries(viper.GetInt("blockstore.s3.max_retries"))
	return cfg
}

func GetAwsAccessKeyID(awsConfig *aws.Config) (string, error) {
	awsCredentials, err := awsConfig.Credentials.Get()
	if err != nil {
		return "", fmt.Errorf("access AWS credentials: %w", err)
	}
	return awsCredentials.AccessKeyID, nil
}

func GetAccount(awsConfig *aws.Config) (string, error) {
	accessKeyID, err := GetAwsAccessKeyID(awsConfig)
	if err != nil {
		return "", err
	}
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return "", fmt.Errorf("get AWS session: %w", err)
	}
	sess.ClientConfig(sts.ServiceName)
	svc := sts.New(sess)

	account, err := svc.GetAccessKeyInfo(&sts.GetAccessKeyInfoInput{
		AccessKeyId: aws.String(accessKeyID),
	})
	if err != nil {
		return "", fmt.Errorf("get access key info for %s: %w", accessKeyID, err)
	}
	return *account.Account, nil
}

func (c *Config) GetBlockstoreType() string {
	return viper.GetString("blockstore.type")
}

func (c *Config) GetBlockAdapterS3Params() (blockparams.S3, error) {
	cfg := c.GetAwsConfig()

	return blockparams.S3{
		AwsConfig:             cfg,
		StreamingChunkSize:    viper.GetInt("blockstore.s3.streaming_chunk_size"),
		StreamingChunkTimeout: viper.GetDuration("blockstore.s3.streaming_chunk_timeout"),
	}, nil
}

func (c *Config) GetBlockAdapterLocalParams() (blockparams.Local, error) {
	localPath := viper.GetString("blockstore.local.path")
	path, err := homedir.Expand(localPath)
	if err != nil {
		return blockparams.Local{}, fmt.Errorf("could not parse blockstore location URI: %w", err)
	}

	return blockparams.Local{Path: path}, err
}

func (c *Config) GetBlockAdapterGSParams() (blockparams.GS, error) {
	return blockparams.GS{
		CredentialsFile: viper.GetString("blockstore.gs.credentials_file"),
		CredentialsJSON: viper.GetString("blockstore.gs.credentials_json"),
	}, nil
}

func (c *Config) GetAuthCacheConfig() authparams.ServiceCache {
	return authparams.ServiceCache{
		Enabled:        viper.GetBool("auth.cache.enabled"),
		Size:           viper.GetInt("auth.cache.size"),
		TTL:            viper.GetDuration("auth.cache.ttl"),
		EvictionJitter: viper.GetDuration("auth.cache.jitter"),
	}
}

func (c *Config) GetAuthEncryptionSecret() []byte {
	secret := viper.GetString("auth.encrypt.secret_key")
	if len(secret) == 0 {
		panic(fmt.Errorf("%w. Please set it to a unique, randomly generated value and store it somewhere safe", ErrMissingSecretKey))
	}
	return []byte(secret)
}

func (c *Config) GetS3GatewayRegion() string {
	return viper.GetString("gateways.s3.region")
}

func (c *Config) GetS3GatewayDomainName() string {
	return viper.GetString("gateways.s3.domain_name")
}

func (c *Config) GetS3GatewayFallbackURL() string {
	return viper.GetString("gateways.s3.fallback_url")
}

func (c *Config) GetListenAddress() string {
	return viper.GetString("listen_address")
}

func (c *Config) GetStatsEnabled() bool {
	return viper.GetBool("stats.enabled")
}

func (c *Config) GetStatsAddress() string {
	return viper.GetString("stats.address")
}

func (c *Config) GetStatsFlushInterval() time.Duration {
	return viper.GetDuration("stats.flush_interval")
}

// GetCommittedTierFSParams returns parameters for building a tierFS.  Caller must separately
// build and populate Adapter.
func (c *Config) GetCommittedTierFSParams() (*pyramidparams.Params, error) {
	adapter, err := factory.BuildBlockAdapter(c)
	if err != nil {
		return nil, fmt.Errorf("build block adapter: %w", err)
	}
	logger := logging.Default().WithField("module", "pyramid")
	return &pyramidparams.Params{
		FSName:             "committed",
		Logger:             logger,
		Adapter:            adapter,
		BlockStoragePrefix: viper.GetString("committed.block_storage_prefix"),
		Local: pyramidparams.LocalDiskParams{
			BaseDir:        viper.GetString("committed.local_cache.dir"),
			AllocatedBytes: viper.GetInt64("committed.local_cache.size_bytes"),
		},
	}, nil
}

func GetMetastoreAwsConfig() *aws.Config {
	cfg := &aws.Config{
		Region: aws.String(viper.GetString("metastore.glue.region")),
		Logger: &LogrusAWSAdapter{},
	}
	if viper.IsSet("metastore.glue.profile") || viper.IsSet("metastore.glue.credentials_file") {
		cfg.Credentials = credentials.NewSharedCredentials(
			viper.GetString("metastore.glue.credentials_file"),
			viper.GetString("metastore.glue.profile"))
	}
	if viper.IsSet("metastore.glue.credentials.access_key_id") {
		cfg.Credentials = credentials.NewStaticCredentials(
			viper.GetString("metastore.glue.credentials.access_key_id"),
			viper.GetString("metastore.glue.credentials.access_secret_key"),
			viper.GetString("metastore.glue.credentials.session_token"))
	}
	return cfg
}

func GetMetastoreHiveURI() string {
	return viper.GetString(MetaStoreHiveURI)
}

func GetMetastoreGlueCatalogID() string {
	return viper.GetString(MetastoreGlueCatalogID)
}
func GetMetastoreType() string {
	return viper.GetString(MetaStoreType)
}

func GetFixedInstallationID() string {
	return viper.GetString("installation.fixed_id")
}
