package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CryptoDepositAddress holds the schema definition for the CryptoDepositAddress entity.
//
// 用途：存储用户加密货币充值地址
// 删除策略：硬删除
// CryptoDepositAddress 使用硬删除，地址状态通过 status 字段管理
type CryptoDepositAddress struct {
	ent.Schema
}

func (CryptoDepositAddress) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "crypto_deposit_addresses"},
	}
}

func (CryptoDepositAddress) Fields() []ent.Field {
	return []ent.Field{
		// 用户关联
		field.Int64("user_id"),

		// 币种和网络信息
		field.String("currency").
			MaxLen(20).
			Comment("币种，如 'USDT', 'BTC', 'ETH'"),
		field.String("network").
			MaxLen(20).
			Comment("区块链网络，如 'TRC20', 'ERC20', 'BEP20'"),

		// 充值地址
		field.String("address").
			MaxLen(128).
			Comment("充值地址"),

		// 服务商信息（可选，用于第三方支付网关）
		field.String("provider_instance_id").
			Optional().
			Nillable().
			MaxLen(64).
			Comment("支付服务商实例 ID"),

		// 地址状态
		field.String("status").
			MaxLen(30).
			Default("active").
			Comment("地址状态：active（活跃）、inactive（停用）、expired（过期）"),

		// 时间戳
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (CryptoDepositAddress) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("crypto_deposit_addresses").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (CryptoDepositAddress) Indexes() []ent.Index {
	return []ent.Index{
		// 唯一约束：同一用户对于同一币种和网络只能有一个充值地址
		index.Fields("user_id", "currency", "network").
			Unique(),
		index.Fields("user_id"),
		index.Fields("address"),
		index.Fields("status"),
	}
}
