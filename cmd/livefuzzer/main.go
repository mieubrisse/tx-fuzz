package main

import (
	"context"
	"crypto/ecdsa"
	crand "crypto/rand"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	txfuzz "github.com/kurtosis-tech/tx-fuzz"
)

const (
	numSpammingThreads = 10
)

var (
	txPerAccount = 1000
	airdropValue = new(big.Int).Mul(big.NewInt(int64(txPerAccount*100000)), big.NewInt(params.GWei))
	corpus       [][]byte

	// TODO Parameterize these instead!!!! This only works with a single genesis
	keys = []string{
		"0xaf5ead4413ff4b78bc94191a2926ae9ccbec86ce099d65aaf469e9eb1a0fa87f",
		"0xe63135ee5310c0b34c551e4683ad926dce90062b15e43275f9189b0f29bc784c",
		"0xc216a7b5048e6ea2437b20bc9c7f9a57cc8aefd5aaf6a991c4db407218ed9e77",
		"0xc29d916f5b6ddd0aa2c827ab7333e40e91fda9ca980332b3c60cae5b7263dae7",
		"0xde1013a3fcdf8b204f90692478254e78126f9763f475cdaba9fc6a1abfb97db3",
		"0xb56fd1fd33b71c508e92bf71fae6e92c76fcd5c0df37a39ff3caa244ddba3c0f",
		"0x10ee939f126a0c5fc3c3cc0241cd945aa88f57eef36bde8707db51ceecfd9134",
		"0x03210fac527544d5b49b89763121b476c4ab66908b345916d6ad2c7740f23803",
		"0x8a6ccbba94844d3951c1e5581df9f8f87de8106b995a3a664d9130a2b72a4b96",
		"0x831a53a7994ac56f6c5d99d6371d7a8686f385995da2592aac82dda8b008b454",
		"0x5a1f678991d52ca86f2b0403727d19887a261199fd88908bde464f7bf13aa50b",
		"0x9823c88ce58b714d576741ae3b97c7f8445209e73596d7588bc5101171e863f4",
		"0x0bd4b408eb720ecc526cf72c88180a60102dd7fd049a5ad04e367c0d9bbc843e",
		"0x4e5ae855d26b0fdc601c8749ad578ba1ddd714997c40e0483479f0f9ff61c51d",
		"0xa0a0419689b82172eb4b9ee2670fef5e1bfa22909b51b7e4f9ab8a9719045a43",
		"0x7be763c508ccfd597ab4854614c55162506d454cd41627c46b264483bb8d2138",
		"0x5d6f77078b8f61d0258737a29c9d8fe347d70af006c109c5df91ae61a85be52b",
		"0x6737e36a661e8de28b5c448063a936bd6c310eab319338f4c503b8f18e935007",
		"0xbd5d6497a71ea419b0d0dc538016c238b8a6a630f6cdda299fcc4ce85858f95b",
		"0x8848fc11b20202e9f702ea05eed08e509eb8d5e32e61454f114bf2085392df75",
		"0x2b7aad673e69b9e4b2d7963ef27e503999f0bd5ff3e63f292e188d2e7df2fe60",
		"0xd2254db4e9d74fd2ab7c296f6bc06ab6beb11d8e3f685a582efc2b5e2cc8f86c",
		"0x5477ebb68a387dc9d4cf2c2b194bed9a027e7f817bd2cac553aca9fe1ec923ad",
		"0xb68a0d9d69df9697ce2c8e4e3a629e7400ddb88a93879a48788c8e8404b2ff90",
		"0xd2db07c60da1bf2048b84c1e09fe4d5bb1b6d0b0eb06bef801e1c2bac1c93d76",
		"0x9e759f9762cb967f96fe437cfa432e2889b2f972ca9f382756efb4998188be12",
		"0x18b886f1e77682ae7a92e9d1c29c13acfb2f493a69723b156510711526654e4f",
		"0x907c2e461495607062e0a7ad8bdb29d7129209ba1accbb478dbd3dee7671a1c8",
		"0xa63e812d650015a9ec0fc94c09b02cc9425e3e197be4a41f1b44a869dd3adace",
		"0xc9a3d46fa54409b795df80a363f43ae31e0c7d15f7d4c5062ce88ae7b78124b8",
		"0x6d41eb903d4f5b21e29a8d8558be7dce002e1b23298d2df7cee4dfdaff5b5980",
		"0xe969e9be3a7e87dc29699f61d5566c79d95803779575df98d054f4207b363333",
		"0xbf82a18972fadc7bf60d8c5bcbd37c8a55fe0cfbb17106e0617ebe6999b2bb61",
		"0x011b3a4adb79e6b372972d5a66ea1acdc44b1ca6ea9985af90fda4a12622926b",
		"0x43034269f49a0963cd45473c22d18693ebbd924f515a49b6fd9190dc96ec60de",
		"0x2a36981f40b25474da25277593836451c8ddc0a4fdd9131cd82dcef003a1c4ce",
		"0x59e57a2b3739c119d94e4e2ecf5c0f8430241e59d27f386f17d050d30f1d5d99",
		"0x8d6eb80f206eec85c773585295f850e159a27ba148360a34bb3355caab17f1b2",
		"0x1491bd992bff53671dd787070fdd54122c395690e248dc4eb32b0fb942a17cc2",
		"0xb7f637ebf0faef160984b039994b75fbec1714128eb1feafe92ce3f7e54bbbae",
		"0xfd0a5e72904e4c1497e315896fe6918c513f76503377005b70cf30dbd705fc50",
		"0x760f74cbbc74c4cdc521abcdbe8ae519091311a1cd3dfd04559848bf94a4af71",
		"0x4bc76ef2f36f8988d1c52eec26be1d31f212781bf918e57406a5e8ad14262c36",
		"0x716cce68bb7dace09415047aff1cf90af99ac6b81c128eb6c80ad9b739c3fc47",
		"0xba9d2c4fed88860a2fb2103590337e71e6720d4323e64dc24f1d9f2f98023c28",
		"0xc9eb9929bec8348030b917fca6e0c4a4b08059c831e9c40250eddce6a942d61f",
		"0xcbd0db2dabc113c02254bc50efd457bcc446338cde9c7f5c93be4f155541bf8e",
		"0x0a6c4e2207afd4c74d78883c4ef8c526dd164b67faf8c8da928d4a7c694fa49c",
		"0xde664e3cde706b7dfc33e1c561e9c63f6223c68668507a04804644aa7f56c8ce",
		"0xbefc32493440df7b3a3951e824570f69b6fe8947e5714ed4da101ddcf33e0f25",
		"0x8971997fb5d1a00599d30af8fe169a47e4d17d3efffefba52839537b89f54bb1",
		"0x5011902bd739d52f5d5c0e2bd382fc81350eb2c4f50aa0f1ce973e5fc41378e6",
		"0x3d371359b160afaeaba364bdcae5a6639699a513198ebb12fb3ee1f215e34e49",
		"0x496a267275fc974aac6d3f46921dd6a08441c9f1dbb9861dc5d8210a1d52053a",
		"0xf2e1e9548e5c15a9d21189ee2766770f0245b238499526fe538b505c7a159774",
		"0x48f59be041558af55f36dfc7ec5a7950d6482d10f8c0977b71a0b86e9c0a4767",
		"0xedeb92913dd239b28906688cae03d3790508c7df9b15fa6ef9abb4042c659985",
		"0xb4094d84262c1039240c21f4d8364b7e802940ba734573e9d8e7566573880c41",
		"0xdec2b48cfcd0273ad1f23404e0f9e9d3fc4fa1721c317d933484b9015858ba5d",
		"0x7dbd4f809ecdaa1ab4dcb40f6a24ddc4561fe264c3fe405e3dfcb7eb44f6275d",
		"0xe3cbb18271f2be064b2148bd15449f2cb92169cecee82ace180e38efef35ab99",
		"0xbfcf51ca0f25962a5a121234999a96818f94580fc7cfddabc0d0c6dbb98ba8a1",
		"0xfb2746579a4129d17c58a796832e30eacc3f329fec12babccc69c49601d93e06",
		"0x3ecffcca6aef54316e12532a4284829334a0d0ea98921ece4b8b296133a7d454",
		"0x56340e56c551897a8e4f6206b4dd9c581c242b39f3417a0460b6b1389c1db1cc",
		"0x2ad703afee0c00cd64f8a309df7caf56db476f50a3a63eb7b645f16afe347670",
		"0x499881041292f414309a062c4a0e50a32886a642ca5b18ce2f33892c01c59b6f",
		"0x169e05cb1c52f31bce923a62c3a54878e5009664ebdbda7d2c9c2eb8f9ef6c60",
		"0xeea4623936f85480e06e6fef37b4b8f6609df6060379c06826b53998c65db8a5",
		"0xd190fb5d73d6637e61219b9fec59d262e78ea2425e1b33ac90913d53f4198f57",
		"0x9887e54d2d5080cb822d36ba321f2f8e94ab86ec41202b32364330a66f3771f3",
		"0x27e029823dc6c29f811586f1c7494da3f98b21b66149b84aef9cecbf1b3b7d84",
		"0x1d13d33d4fb6fece8905aa1ca88b12b13d6657c0088e4aef925d9b841c8bd04e",
		"0x917b800fe9c34bb01f64466d72d16b17a1a8ee259fe75388728732fbd85efc56",
		"0xbdc439d615a794cd56f1e8db57443ecfb7a3aab3007b577c1384c326796e416e",
		"0x56f007b4a793aa597f2931918041d15e49c0f0df13d70e943abb1dd55443bdde",
		"0x8cbf6dd9f811bd995360e07cd473b03557b26edaa08c74a2b214b51a62a86add",
		"0x8d995b221fde9d68f694d67387fa5f7ba96cbf214859afdee35ee1259f40258a",
		"0xbfda7549339ec65a6194c590adae05afa8b151ac09e45d47089909baa2e85d0f",
		"0x327e90aacbb89dc399c16dda9bf9c678c34706a92a12f916937d51f06f7b77eb",
		"0xf209378bcbc64c73e8b7f79b458862fa5352ef98e3be6f35b939da11585f88bd",
		"0x48855ae3f55f7541100460044bdffff68d385c037e5d95467ecc7e9bf94717b4",
		"0xa8b35150d825158df3582e93b59697267da48e9c346fa0201b73740d494066c1",
		"0x39d071a2992e1f9cae76f063f6ac43a8c391f48e6dee0a747cb96fc484c8b1ad",
		"0x44f423e403ae230485e9252adad3a7919b15929264e74b153cc71d82e4aa4092",
		"0x0baa66376068c94bb0584b6ffd546b920fe3800bdf0738983b4664936bb77ab5",
		"0x16accf976a71e9ec5a529d4664adc78a3ddf54b8e5c9515c9b5cae0c510f84c5",
		"0x1edc90ec503856ab0ad0d0f94b23f32ed4fd4e0f40b135f742954e045d556cc9",
		"0x3349c201895a20ee27559947736cefaff2c8ea4e4f4596af993a32f96d574c7c",
		"0xe43760622c3706049c2e5f83286dfba2560f7f435f9a84c4a02580ab74ddcd3b",
		"0xcb5b181c33eb5799d13404b3aca7636f4b1754b1edfad6e032031ccef08c0a9b",
		"0xb68f242eb49ecb58129a46ce18f118a1da76f75753bf0b3c955dea35ca453b76",
		"0x88b6798175f3fdc2f6fb9a14c5d0223ce7f20f63bee3c37bcc3cdb19ef15314c",
		"0x8d1859fa479851868ce2c9e364402d393f80f7092aaa56b81669b110052d89c4",
		"0xaed286129bfe7b12eab109f95241eb00d869e951481077ff776169f2bdba5826",
		"0xd7a0a2519649a7ebd7e343291b245411e1d9280755ca844f80499b02ada3cc8e",
		"0xb72018ac17ae122c7af6d27e8bd10646980fa28dc31677de31ad4da676d93590",
		"0x6c8bdac67bcefc51948fa2312d0b96d63b50d64a1c8fc86c10524b5b0d0065bc",
		"0xe4dac2bcaed15966dade7969b3c846218333ea30aac8b40bb1f9aefaf450bf7f",
		"0xa58a4dabe4e61062381f3c5cfbdf9475da4d7753b95c2ceda3146a839317ef22",
	}
	addrs = []string{
		"0x6177843db3138ae69679A54b95cf345ED759450d",
		"0x687704DB07e902e9A8B3754031D168D46E3D586e",
		"0x15E6A5a2E131Dd5467fA1FF3aCD104F45EE5940B",
		"0x80C4C7125967139AcAa931Ee984A9Db4100e0F3B",
		"0xD08A63244fCD28b0Aec5075052cdCe31BA04fEAD",
		"0x0b06ef8be65fCdA88F2dbaE5813480f997eE8E35",
		"0x1cB96C5809dA5977f99f69c11eE58BaE5711c5f1",
		"0x2AA48eE899410a6d97C01B0bb0eeaf1771CC435B",
		"0x07B9d920Dd8E8d83Dc1125c94fc0b3cdCDf602Fb",
		"0xfCB6E353AD4F79245C7cB704ABCfFe2F48684241",
		"0x0D3dE4256d6322683FdEa9EE23765ccBfCB83DA4",
		"0x6021752d8D9B2f221d4fEa4349Dea34dDBcfcE50",
		"0x61e296D527eDC89e831Cf593eC341f16197eEAFB",
		"0xcf7317Ee7A3B497eCf634B94BfF60FF91B925747",
		"0x7e7b519dF31f77CeD83eEa1B16AEDB6dcb0F0b24",
		"0x88a075e0fB1C9309A200A8bF0A88B214BF7Ceb8d",
		"0xc8D7cfb58f3aC02568E6505Bf3fB5EB6F0807039",
		"0xE0132E8D7B1b766E0adE5543d6C6C0b2D5A2F01d",
		"0xeb674c0411Db79654afDc1e131F3B6e734baEE6C",
		"0xdC07c60993cf689438b8c85f86B0Ed938dCA77eA",
		"0x110dDC93Db59eD31A03518510221EC2f35d28f2f",
		"0xb599A876AAac824cfcE21BDf15627c9Fd8634C30",
		"0xd36e5540Dd71aCBd6416d60252C4d7C34a3C8245",
		"0x3adECa35af56206A74987A8Fe13C669365c770cf",
		"0xD77b95acd12f7B4b5692B55717B7bbCA11651954",
		"0xf388bF5766b5ed5d4E1CBf15772e677DBFa80b00",
		"0x35d4996296E58560e6Ef47787d51b55f1E2Bd92A",
		"0xA4c3b77b898E53D6095F11c53A1cE272CFF9af31",
		"0x6E84f6113fc1919714F0266705813fB81a17181F",
		"0xe9AE1A806004e1452baae0493920815AAdD84798",
		"0xfe1905d8EBd20E037274eef441283C811Ea82C16",
		"0x6aDecE88E477F53a143a4C29D97940DF2Ec768e0",
		"0x0d34d140a7376892C4593fcEA3AE26F5d6F202d7",
		"0xd1c7fa75b9BC55d041fcdf215f3E3A351C9F9edC",
		"0x418EBe350A8C6387BF5E42f3502742Af8e0781f1",
		"0x84914D2770c711D27888c775c547b1D933B48C47",
		"0x8f51E560B85eDF2E653c689c4e9FAC02CE0556B8",
		"0xEE2503205C24Dc66346E356f13f333FB8782d358",
		"0x096bA6C59Bd667A0fEa9A356BCC988E4D9F2D8eB",
		"0xDA0adcE4f1Dc7DEbE7B2b52e8fe9ace6C7ea9C66",
		"0xaF7d412AEAB7525C0541dc3AA6C1085CFb8C9099",
		"0x3cF8c0d567261EaF4aC0872D33A9F48AF361769f",
		"0x4779242587ba9e828999249eAdD82984430f4843",
		"0xea531cfE2dE357ECff3855B88dbd07f60b03cDca",
		"0xd00B5F53Ea2a66ad33c3feE304bB22857dfb8A87",
		"0x7EAd29F6616F78F21A951c9686dD257bE7b8EFE4",
		"0xD503c13Ee55C1ea128357d4018Ec58d0D5E5C3db",
		"0x4AC670d8760faf780468638eF80034876ed8918D",
		"0x24fFb8C97CE443f8d3265bA3316DEfCfC07C659c",
		"0x0c5cAfc547ab98c9CEAA1C07Fdd6Bf7820aEB954",
		"0xdB8d964741C53E55dF9C2d4e9414c6C96482874e",
		"0xBA85bb35ae6Ff7a34745993fcF92b9afD34124f1",
		"0x58871015F5a2D3948264F7c16Ad194c80FfD531D",
		"0x2A90Af45dF70B0031F218cc122598dDf3E10469f",
		"0x761BbAaEA6ceB265F5262C3B559aDc2AD3ED2F09",
		"0xdFe86f51C5E603F1420d1F0Ab366bd3BFe23D2A7",
		"0xd616547158B05AB5079106DC0336d72763A72871",
		"0xdC68CD278cB7F5f666CE7b0a3a214A8540ed4dFa",
		"0x11F8107da05B6905e8cC0227CA3B0c6Eb764FaC0",
		"0x04DA906545679850A7Ee0Ef6836e183031BeDC88",
		"0x8Bdc25c43c010fD3db6281fCd8f7A0BeD18838e3",
		"0xaf16F746b8A834A383fd0597D941FEE52B7791Eb",
		"0x0C5C736600f8ea58cCb89aa72e3f3634651FD551",
		"0x6F475E0f0E9EdA58556FDdc04De9B1a9B6A4CFb4",
		"0x9b2e76498a695c4DC7D0890069Cffa84a9581D24",
		"0xE2d2B2069f4a54Fcc171223FF0c17ADbd743c285",
		"0x386bD49f04322544F3c7178Fa5Ae1a24B947B454",
		"0x00af839c3Fc067FAfc2e0a205858D6957F0dd18D",
		"0xeBb6d32A650Afa9221B55a11c6A6de52b6F07Cd7",
		"0x011d26A3A9ADc9203c8943A6A77aA8657Af52420",
		"0x9C85bc61a89fb5ABd957E6c819c653fc1AA0D11b",
		"0xBD8e8435b7897D87cf7ceDb5Cf8c5dD865dBF720",
		"0xAdeBeE2e3ff041078b62380D001C6E51B4F15598",
		"0x71e94C459c9F05085Fc0d34B5f21e648e05Dc6b3",
		"0x7C1fE317dB82c9298b87C56c3194178271B621e1",
		"0xE069d1c9aBf5127BDc3A164fb93B96Bfa9F74CE0",
		"0xB9bBdDd1EB6eF8FB1BDC6a853D5Ad7486a9487DD",
		"0xA804387CdAf986d45831E8074eFb2115af053F7a",
		"0xf23501D784A041fC911b4c86c2bfb1F63EC170Ea",
		"0x3928Be2a7058088313C0Fb3294014e88A3C5ed4a",
		"0x196AA07204141478459C14106ef5E5282efE9957",
		"0x763Cbf89560e2da270000822aBDa9584db693FA3",
		"0x7feAeA0ff70FfC9eEC2104F57f7136aFF4dEa680",
		"0xE5466AACd9Dd6d3Bb35060A1CCC76a438dE88cA1",
		"0xf670980415CFE8C4f8d10645Ecf974c9A2Fea00E",
		"0xa29115BCe7829ffdD989B7CF1BdD1eAc06A2cb36",
		"0x8F528aA67dC1846C893465fA1C8c26556bc5FE19",
		"0x4Dc4ec6ac43C8c45777292Db987203C0248e17b7",
		"0x0D2F39f251CB547CbA567A31E5e9F93C19dfFA85",
		"0x9eB31FB94ce5111e2a04cb9D156b513887CCBd00",
		"0x04B88eF83f8c41b1465D360A1E82f07Ae190892a",
		"0xAF23e04B04fbE15630eadd32a6f27a5a65EA554a",
		"0x746CdFF371e3f1E905b3Ac52280078bAc2dec7dd",
		"0xc33E5155BDBf1a0A7CeB1b80F8586c5cDA5C3781",
		"0xe7FdEF5f5219068f3D0F88a7445005574C662798",
		"0xf0A81A63C5E09B0BD08e027DE48058E377D3732d",
		"0x9878ab34Dc3B4a63c80fDb733491472c11d59A56",
		"0x912859bEBAE3086aC7A062DEe5D68AA8eD2D71EC",
		"0x5A0B737Ed85049410E5ea61f444D07d5c8C0359F",
		"0x305a5dfd46e6128AbCe28c03B3ad971f4e4915ff",
	}
)

func main() {
	// eth.sendTransaction({from:personal.listAccounts[0], to:"0xb02A2EdA1b317FBd16760128836B0Ac59B560e9D", value: "100000000000000"})
	if len(os.Args) < 3 {
		panic(fmt.Sprintf("Usage: %v node_ip:rpc_port command [no-al]", os.Args[0]))
	}

	rpcUrl := os.Args[1]
	command := os.Args[2]

	accesslist := true
	if len(os.Args) == 3 && os.Args[3] == "no-al" {
		accesslist = false
	}

	switch command {
	case "airdrop":
		panic("TODO Can't airdrop on generic networks until the faucet account is parameterized")
		airdrop(rpcUrl, airdropValue)
	case "spam":
		// The private keys of the addresses that will send transactions
		commaSeparatedPrivateKeys := os.Args[3]
		// The addresses that the private keys correspond to
		commaSeparatedAddresses := os.Args[4]
		SpamTransactions(rpcUrl, commaSeparatedPrivateKeys, commaSeparatedAddresses, false, accesslist)
	case "corpus":
		// The private keys of the addresses that will send transactions
		commaSeparatedPrivateKeys := os.Args[3]
		// The addresses that the private keys correspond to
		commaSeparatedAddresses := os.Args[4]
		cp, err := readCorpusElements(os.Args[5])
		if err != nil {
			panic(err)
		}
		corpus = cp
		SpamTransactions(rpcUrl, commaSeparatedPrivateKeys, commaSeparatedAddresses, true, accesslist)
	case "unstuck":
		unstuckTransactions(rpcUrl)
	case "send":
		send(rpcUrl)
	default:
		fmt.Println("unrecognized option")
	}
}

func SpamTransactions(rpcUrl string, commaSeparatedPrivateKeys string, commaSeparatedAddresses string, fromCorpus bool, accessList bool) {
	backend, _ := getRealBackend(rpcUrl)

	privateKeyStrs := strings.Split(commaSeparatedPrivateKeys, ",")
	addressStrs := strings.Split(commaSeparatedAddresses, ",")

	privateKeys := []*ecdsa.PrivateKey{}
	for _, keyStr := range privateKeyStrs {
		key := crypto.ToECDSAUnsafe(common.FromHex(keyStr))
		privateKeys = append(privateKeys, key)
	}

	addresses := []common.Address{}
	for _, addressStr := range addressStrs {
		addr := common.HexToAddress(addressStr)
		addresses = append(addresses, addr)
	}

	waitgroup := sync.WaitGroup{}
	for i := 0; i < numSpammingThreads; i++ {
		waitgroup.Add(1)
		go func() {
			var f *filler.Filler
			if fromCorpus {
				elem := corpus[rand.Int31n(int32(len(corpus)))]
				f = filler.NewFiller(elem)
			} else {
				rnd := make([]byte, 10000)
				crand.Read(rnd)
				f = filler.NewFiller(rnd)
			}
			SendBaikalTransactions(backend, privateKeys, f, addresses, accessList)
			waitgroup.Done()
		}()
	}
	waitgroup.Wait()
}

// Repeatedly sends transactions from a random source to a random destination
func SendBaikalTransactions(client *rpc.Client, keys []*ecdsa.PrivateKey, f *filler.Filler, addresses []common.Address, al bool) {
	backend := ethclient.NewClient(client)

	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	for {
		// Pick a random source to send ETH from
		idx := rand.Intn(len(keys))
		key := keys[idx]
		srcAddr := addresses[idx]

		nonce, err := backend.NonceAt(context.Background(), srcAddr, big.NewInt(-1))
		if err != nil {
			panic(err)
		}
		tx, err := txfuzz.RandomValidTx(client, f, srcAddr, nonce, nil, nil, al)
		if err != nil {
			fmt.Printf("An error occurred sending transaction from address '%v': %v\n", srcAddr.String(), err)
			continue
		}
		signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainid), key)
		if err != nil {
			panic(err)
		}
		err = backend.SendTransaction(context.Background(), signedTx)
		if err == nil {
			nonce++
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		if _, err := bind.WaitMined(ctx, backend, signedTx); err != nil {
			fmt.Printf("Wait mined failed: %v\n", err.Error())
		}
		cancel()
	}
}

func unstuckTransactions(rpcUrl string) {
	backend, _ := getRealBackend(rpcUrl)
	client := ethclient.NewClient(backend)
	// Now let everyone spam baikal transactions
	var wg sync.WaitGroup
	wg.Add(len(keys))
	for i, key := range keys {
		go func(key, addr string) {
			sk := crypto.ToECDSAUnsafe(common.FromHex(key))
			unstuck(sk, client, common.HexToAddress(addr), common.HexToAddress(addr), common.Big0, nil)
			wg.Done()
		}(key, addrs[i])
	}
	wg.Wait()
}

func readCorpusElements(path string) ([][]byte, error) {
	stats, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	corpus := make([][]byte, 0, len(stats))
	for _, file := range stats {
		b, err := ioutil.ReadFile(fmt.Sprintf("%v/%v", path, file.Name()))
		if err != nil {
			return nil, err
		}
		corpus = append(corpus, b)
	}
	return corpus, nil
}

func send(rpcUrl string) {
	backend, _ := getRealBackend(rpcUrl)
	client := ethclient.NewClient(backend)
	to := common.HexToAddress(txfuzz.ADDR)
	sk := crypto.ToECDSAUnsafe(common.FromHex(txfuzz.SK2))
	value := new(big.Int).Mul(big.NewInt(100000), big.NewInt(params.Ether))
	sendTx(sk, client, to, value)
}

func airdrop(rpcUrl string, value *big.Int) {
	client, sk := getRealBackend(rpcUrl)
	backend := ethclient.NewClient(client)
	sender := common.HexToAddress(txfuzz.ADDR)
	var tx *types.Transaction
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	for _, addr := range addrs {
		nonce, err := backend.PendingNonceAt(context.Background(), sender)
		if err != nil {
			panic(err)
		}
		to := common.HexToAddress(addr)
		gp, _ := backend.SuggestGasPrice(context.Background())
		tx2 := types.NewTransaction(nonce, to, value, 21000, gp, nil)
		signedTx, _ := types.SignTx(tx2, types.LatestSignerForChainID(chainid), sk)
		backend.SendTransaction(context.Background(), signedTx)
		tx = signedTx
	}
	// Wait for the last transaction to be mined
	bind.WaitMined(context.Background(), backend, tx)
}

func getRealBackend(address string) (*rpc.Client, *ecdsa.PrivateKey) {
	// eth.sendTransaction({from:personal.listAccounts[0], to:"0xb02A2EdA1b317FBd16760128836B0Ac59B560e9D", value: "100000000000000"})

	sk := crypto.ToECDSAUnsafe(common.FromHex(txfuzz.SK))
	if crypto.PubkeyToAddress(sk.PublicKey).Hex() != txfuzz.ADDR {
		panic(fmt.Sprintf("wrong address want %s got %s", crypto.PubkeyToAddress(sk.PublicKey).Hex(), txfuzz.ADDR))
	}
	cl, err := rpc.Dial(address)
	if err != nil {
		panic(err)
	}
	return cl, sk
}

func sendTx(sk *ecdsa.PrivateKey, backend *ethclient.Client, to common.Address, value *big.Int) {
	sender := common.HexToAddress(txfuzz.ADDR)
	nonce, err := backend.PendingNonceAt(context.Background(), sender)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Nonce: %v\n", nonce)
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	gp, _ := backend.SuggestGasPrice(context.Background())
	tx := types.NewTransaction(nonce, to, value, 500000, gp, nil)
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainid), sk)
	backend.SendTransaction(context.Background(), signedTx)
}

func unstuck(sk *ecdsa.PrivateKey, backend *ethclient.Client, sender, to common.Address, value, gasPrice *big.Int) {
	blocknumber, err := backend.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}
	nonce, err := backend.NonceAt(context.Background(), sender, big.NewInt(int64(blocknumber)))
	if err != nil {
		panic(err)
	}
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Acc: %v Nonce: %v\n", sender, nonce)
	if gasPrice == nil {
		gasPrice, _ = backend.SuggestGasPrice(context.Background())
	}
	tx := types.NewTransaction(nonce, to, value, 21000, gasPrice, nil)
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainid), sk)
	backend.SendTransaction(context.Background(), signedTx)
}
