
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS `ToDo` (
		`ID` bigint(20) NOT NULL AUTO_INCREMENT,
		`Title` varchar(200) DEFAULT NULL,
		`Description` varchar(1024) DEFAULT NULL,
		`Reminder` timestamp NULL DEFAULT NULL,
		`Status` varchar(200) DEFAULT 'progress',
		`EstimatedTimeOfCompletion` timestamp NULL DEFAULT CURRENT_TIMESTAMP, 
		`ActualTimeOfCompletion` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (ID),
		UNIQUE KEY ID_UNIQUE (ID),
		UNIQUE KEY TITLE_UNIQUE (Title)); 


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE `ToDo`;

