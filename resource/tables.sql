CREATE TABLE IF NOT EXISTS `tbl_alloc_info` (
    `service_name`        VARCHAR(64)     NOT NULL PRIMARY KEY,
    `last_alloc_value`    BIGINT UNSIGNED NOT NULL DEFAULT '0',
	`data_version`        BIGINT UNSIGNED NOT NULL DEFAULT '0'
) ENGINE = InnoDB CHARACTER SET = utf8mb4;
