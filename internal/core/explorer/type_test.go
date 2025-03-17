package explorer

import (
	"context"
	ent "git.web3gate.ru/web3/nft/GraphForge/internal/entity"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
	"testing"
)

var upstream = struct {
	url string
	net string
}{
	url: "https://b.dev.web3gate.ru:32443/045320f8-912e-4a30-a8c3-980c809aeb17",
	net: "sepolia",
}

func TestExplorer_Type(t *testing.T) {
	client, err := ethclient.Dial(upstream.url)
	if err != nil {
		panic(err)
	}

	var e = NewTokenDetector(map[string]*ethclient.Client{upstream.net: client}, zap.NewNop())

	type args struct {
		ctx        context.Context
		deployment *ent.Contract
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "erc20",
			args: args{
				ctx: context.Background(),
				deployment: &ent.Contract{
					Network: upstream.net,
					ChainID: 11155111,
					Address: "0x810A3B22c91002155d305C4Ce032978E3A97F8c4",
				},
			},
			want:    ent.UnknownType,
			wantErr: false,
		},
		{
			name: "erc721",
			args: args{
				ctx: context.Background(),
				deployment: &ent.Contract{
					Network: upstream.net,
					ChainID: 11155111,
					Address: "0x1238536071E1c677A632429e3655c799b22cDA52",
				},
			},
			want:    ent.ERC721Type,
			wantErr: false,
		},
		{
			name: "erc1155-1",
			args: args{
				ctx: context.Background(),
				deployment: &ent.Contract{
					Network: upstream.net,
					ChainID: 11155111,
					Address: "0x0E9b80778d0E4c9D701E55C07a2c0F154263bB19",
				},
			},
			want:    ent.ERC1155Type,
			wantErr: false,
		},
		{
			name: "erc1155-2",
			args: args{
				ctx: context.Background(),
				deployment: &ent.Contract{
					Network: upstream.net,
					ChainID: 11155111,
					Address: "0x26ddE3091C5B372eDe000ECfe4eB91dc3D80C32C",
				},
			},
			want:    ent.ERC1155Type,
			wantErr: false,
		},
		{
			name: "erc1155-3",
			args: args{
				ctx: context.Background(),
				deployment: &ent.Contract{
					Network: upstream.net,
					ChainID: 11155111,
					Address: "0x6df08BFB7f0B9C40CA36A96c477ec7114825B9eb",
				},
			},
			want:    ent.ERC1155Type,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := e.Type(tt.args.ctx, tt.args.deployment)
			if (err != nil) != tt.wantErr {
				t.Errorf("Type() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Type() got = %v, want %v", got, tt.want)
			}
		})
	}
}
