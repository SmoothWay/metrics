package crypt

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed test-public.pem
var publicKeyRaw []byte

//go:embed test-private.pem
var privateKeyRaw []byte

func TestEncryption(t *testing.T) {
	publicKey, err := DecodeKey(publicKeyRaw)
	require.NoError(t, err)

	privateKey, err := DecodeKey(privateKeyRaw)
	require.NoError(t, err)

	testCases := []struct {
		name string
		src  []byte
		want []byte
	}{
		{
			name: "Positive case #1",
			src:  []byte("not encoded string"),
		},
		{
			name: "Large data",
			src: []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Ut ipsum nisi, interdum eget iaculis sit amet, hendrerit non risus. Sed aliquet, mauris eget rutrum egestas, orci lectus mollis nibh, sed ornare metus erat eu ex. Pellentesque id ante eu tellus egestas pellentesque at eu leo. Morbi egestas tortor sit amet augue tincidunt aliquet. Curabitur vitae metus vel metus eleifend blandit sed non tortor. Sed et erat quam. Vestibulum nunc tortor, tincidunt ut finibus nec, scelerisque et turpis. Aenean placerat, neque eu efficitur dictum, nunc arcu maximus mauris, eu commodo massa risus et quam. Sed et gravida dui. Nunc nulla tortor, interdum et nisi et, malesuada sollicitudin purus. Ut sollicitudin ultrices porta. Vivamus eu placerat elit, sit amet tincidunt nisi. Donec finibus sollicitudin quam, a consectetur sapien tristique in. Nulla porta lorem ut libero fermentum laoreet. Aliquam dapibus porttitor neque eleifend suscipit.\n" +
				"In nec ipsum fermentum diam rutrum ultricies vel eu mi. Etiam vitae massa neque. Sed cursus eros lorem, in gravida ipsum consectetur vitae. Praesent sit amet leo quam. Praesent orci tortor, porttitor ut diam sit amet, porta volutpat lorem. Morbi eget tincidunt ante. Donec rutrum nec urna sed dignissim. Integer eleifend cursus tellus, vel sagittis tellus euismod eu.\n" +
				"Phasellus eu euismod dui, eleifend finibus orci. Sed in dui ornare, efficitur sem porta, congue magna. Phasellus a laoreet nisl. Ut luctus risus urna, nec sodales lacus ultricies eget. In eleifend, leo vel aliquam maximus, nisl nunc posuere elit, quis iaculis nulla urna ut justo. Aliquam et nunc interdum, egestas libero id, consectetur est. Praesent laoreet ut mi eget sodales.\n" +
				"Ut faucibus maximus neque id eleifend. Etiam sed est posuere lectus volutpat sollicitudin et vitae dolor. Mauris rhoncus sed lectus et bibendum. Sed ut scelerisque ligula. Pellentesque tincidunt ante vitae ipsum malesuada tristique. Integer ut tincidunt est, quis suscipit turpis. Cras rutrum lorem dui, eu accumsan quam ornare vel. Nullam commodo non ex non convallis. Nulla venenatis nunc dui, ut rutrum ligula dapibus vitae. Vestibulum malesuada, diam at posuere tempor, arcu ligula hendrerit orci, a pharetra mi lacus vel libero. Proin vitae lobortis nibh. Phasellus efficitur aliquam erat, et posuere leo sagittis ut. Donec vitae purus sed turpis scelerisque accumsan at vel elit. Sed mi orci, varius quis justo a, ullamcorper sodales sapien. Aliquam nec massa quis eros egestas tristique vitae quis nibh. Donec sit amet est quis enim tempor fringilla in ut orci.\n" +
				"Aenean consequat, ipsum id tempus auctor, libero magna volutpat enim, quis ornare eros lectus nec dui. Aenean eu eros id diam porta mollis. Aenean faucibus justo vitae diam egestas, non bibendum nisl tristique. Fusce ultricies ullamcorper ligula fringilla dignissim. Nam sit amet ipsum non purus molestie eleifend vel ac purus. Donec a ante libero. Vivamus volutpat turpis quis diam auctor viverra. Aliquam id imperdiet purus, ut ornare risus. Pellentesque in ipsum cursus, faucibus urna a, posuere risus.\n" +
				"Duis lectus ligula, mollis in augue sed, ornare commodo leo. Aliquam arcu enim, scelerisque eu libero in, ultricies consectetur lorem. Ut ac risus vel elit vulputate fermentum id ut dui. Fusce aliquet quis augue ut cursus. Donec suscipit, enim in venenatis sodales, purus lorem elementum odio, non blandit lorem arcu a magna. Proin ullamcorper massa sit amet risus gravida, nec tincidunt nulla congue. Vivamus mattis mollis sodales. Aenean ac augue congue, mollis sapien quis, sollicitudin diam. Aliquam erat volutpat. Nulla nec arcu non est bibendum venenatis. Nam iaculis varius sodales. Nunc nec tellus non sapien semper convallis. Vivamus imperdiet tortor ac quam viverra, sit amet fermentum ligula dictum. Curabitur ac mauris suscipit, congue mi vel, sagittis justo. Morbi non posuere nisl, a eleifend nibh.\n" +
				"Sed scelerisque suscipit purus, id iaculis nisl mollis eu. Proin auctor egestas velit, feugiat ullamcorper metus consectetur non. Donec sed ornare metus, non laoreet diam. Curabitur odio ante, dignissim eget ex ac, aliquet euismod erat. Mauris rutrum ante ex, vel placerat elit congue at. Pellentesque laoreet ut purus non sollicitudin. Maecenas pharetra eget dolor nec fermentum. Vestibulum at quam tristique, tincidunt ex ac, molestie felis. Suspendisse tincidunt arcu sit amet lorem aliquam ultricies. Proin non placerat orci. Nam viverra tellus sed molestie aliquet.\n" +
				"Nam tincidunt pellentesque arcu in congue. Curabitur eget magna viverra, congue risus auctor, pretium felis. Sed vulputate fringilla magna at bibendum. Nulla sed eleifend enim, non luctus elit. Sed et nisi convallis justo sodales tempus. Nulla quis congue urna. Mauris vitae dictum risus. Nulla euismod rutrum nunc, non maximus ante vestibulum vitae. Integer lacinia sapien est. Nam suscipit nibh rutrum consequat facilisis. Nam nec metus placerat, euismod erat nec, placerat nibh. Duis sed consequat nunc. Quisque tincidunt, enim et sodales congue, magna purus mi."),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := Encrypt(tc.src, publicKey)
			require.NoError(t, err)
			assert.NotEqual(t, tc.src, encoded)

			decoded, err := Decrypt(encoded, privateKey)
			require.NoError(t, err)
			assert.Equal(t, tc.src, decoded)
		})
	}
}
