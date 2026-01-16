-- =====================================================
-- Migration: 000001_init_schema
-- Description: Create core banking tables (accounts, entries, transfers)
-- =====================================================

-- accounts: 银行账户表
-- 每个账户有唯一ID、所有者、余额和货币类型
CREATE TABLE "accounts" (
    "id"         BIGSERIAL PRIMARY KEY,              -- 自增主键
    "owner"      VARCHAR(255) NOT NULL,              -- 账户所有者(用户名)
    "balance"    BIGINT NOT NULL DEFAULT 0,          -- 账户余额(单位:分,避免浮点数精度问题)
    "currency"   VARCHAR(3) NOT NULL,                -- 货币类型 (USD, EUR, CNY)
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- 创建时间(带时区)
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()  -- 更新时间
);

-- 索引: 按所有者查询账户列表
CREATE INDEX "idx_accounts_owner" ON "accounts" ("owner");

-- 唯一约束: 同一用户同一货币只能有一个账户
CREATE UNIQUE INDEX "idx_accounts_owner_currency" ON "accounts" ("owner", "currency");

-- =====================================================

-- entries: 账目记录表
-- 记录每一笔资金变动(入账/出账)
CREATE TABLE "entries" (
    "id"         BIGSERIAL PRIMARY KEY,
    "account_id" BIGINT NOT NULL,                    -- 关联的账户ID
    "amount"     BIGINT NOT NULL,                    -- 金额(正数=入账, 负数=出账)
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- 外键约束: 确保账户存在
    CONSTRAINT "fk_entries_account"
        FOREIGN KEY ("account_id")
        REFERENCES "accounts" ("id")
        ON DELETE CASCADE                            -- 删除账户时级联删除记录
);

-- 索引: 按账户ID查询交易历史
CREATE INDEX "idx_entries_account_id" ON "entries" ("account_id");

-- =====================================================

-- transfers: 转账记录表
-- 记录账户间的转账操作
CREATE TABLE "transfers" (
    "id"              BIGSERIAL PRIMARY KEY,
    "from_account_id" BIGINT NOT NULL,               -- 转出账户
    "to_account_id"   BIGINT NOT NULL,               -- 转入账户
    "amount"          BIGINT NOT NULL,               -- 转账金额(必须为正数)
    "created_at"      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- 外键约束
    CONSTRAINT "fk_transfers_from_account"
        FOREIGN KEY ("from_account_id")
        REFERENCES "accounts" ("id"),

    CONSTRAINT "fk_transfers_to_account"
        FOREIGN KEY ("to_account_id")
        REFERENCES "accounts" ("id"),

    -- 检查约束: 金额必须为正数
    CONSTRAINT "chk_transfers_amount_positive"
        CHECK ("amount" > 0)
);

-- 索引: 按转出/转入账户查询
CREATE INDEX "idx_transfers_from_account_id" ON "transfers" ("from_account_id");
CREATE INDEX "idx_transfers_to_account_id" ON "transfers" ("to_account_id");

-- 复合索引: 同时按双方账户查询
CREATE INDEX "idx_transfers_from_to" ON "transfers" ("from_account_id", "to_account_id");

-- =====================================================
-- 添加表注释
COMMENT ON TABLE "accounts" IS '银行账户表';
COMMENT ON TABLE "entries" IS '账目记录表(入账/出账)';
COMMENT ON TABLE "transfers" IS '转账记录表';

COMMENT ON COLUMN "accounts"."balance" IS '余额(单位:分)';
COMMENT ON COLUMN "entries"."amount" IS '金额(正数入账,负数出账)';
COMMENT ON COLUMN "transfers"."amount" IS '转账金额(必须为正数)';
