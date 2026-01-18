-- =====================================================
-- Migration: 000001_init_schema
-- Description: Create core banking tables (accounts, entries, transfers)
-- Database: MySQL 8.0+
-- =====================================================

-- accounts: 银行账户表
-- 每个账户有唯一ID、所有者、余额和货币类型
CREATE TABLE `accounts` (
    `id`         BIGINT AUTO_INCREMENT PRIMARY KEY,   -- 自增主键
    `owner`      VARCHAR(255) NOT NULL,               -- 账户所有者(用户名)
    `balance`    BIGINT NOT NULL DEFAULT 0,           -- 账户余额(单位:分,避免浮点数精度问题)
    `currency`   VARCHAR(3) NOT NULL,                 -- 货币类型 (USD, EUR, CNY)
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP -- 更新时间
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='银行账户表';

-- 索引: 按所有者查询账户列表
CREATE INDEX `idx_accounts_owner` ON `accounts` (`owner`);

-- 唯一约束: 同一用户同一货币只能有一个账户
CREATE UNIQUE INDEX `idx_accounts_owner_currency` ON `accounts` (`owner`, `currency`);

-- =====================================================

-- entries: 账目记录表
-- 记录每一笔资金变动(入账/出账)
CREATE TABLE `entries` (
    `id`         BIGINT AUTO_INCREMENT PRIMARY KEY,
    `account_id` BIGINT NOT NULL COMMENT '关联的账户ID',
    `amount`     BIGINT NOT NULL COMMENT '金额(正数=入账, 负数=出账)',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- 外键约束: 确保账户存在
    CONSTRAINT `fk_entries_account`
        FOREIGN KEY (`account_id`)
        REFERENCES `accounts` (`id`)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='账目记录表(入账/出账)';

-- 索引: 按账户ID查询交易历史
CREATE INDEX `idx_entries_account_id` ON `entries` (`account_id`);

-- =====================================================

-- transfers: 转账记录表
-- 记录账户间的转账操作
CREATE TABLE `transfers` (
    `id`              BIGINT AUTO_INCREMENT PRIMARY KEY,
    `from_account_id` BIGINT NOT NULL COMMENT '转出账户',
    `to_account_id`   BIGINT NOT NULL COMMENT '转入账户',
    `amount`          BIGINT NOT NULL COMMENT '转账金额(必须为正数)',
    `created_at`      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- 外键约束
    CONSTRAINT `fk_transfers_from_account`
        FOREIGN KEY (`from_account_id`)
        REFERENCES `accounts` (`id`),

    CONSTRAINT `fk_transfers_to_account`
        FOREIGN KEY (`to_account_id`)
        REFERENCES `accounts` (`id`),

    -- 检查约束: 金额必须为正数 (MySQL 8.0.16+)
    CONSTRAINT `chk_transfers_amount_positive`
        CHECK (`amount` > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='转账记录表';

-- 索引: 按转出/转入账户查询
CREATE INDEX `idx_transfers_from_account_id` ON `transfers` (`from_account_id`);
CREATE INDEX `idx_transfers_to_account_id` ON `transfers` (`to_account_id`);

-- 复合索引: 同时按双方账户查询
CREATE INDEX `idx_transfers_from_to` ON `transfers` (`from_account_id`, `to_account_id`);
