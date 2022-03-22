package envsetup

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetupInvitedUser(t *testing.T) {
	base64Cred := "eyJ1cmwiOiJodHRwOi8vbG9jYWxob3N0OjgwOTAvIiwiYWNjZXNzVG9rZW4iOiJleUoyWlhJaU9pSXlJaXdpZEhsd0lqb2lTbGRVSWl3aVlXeG5Jam9pVWxNeU5UWWlMQ0pyYVdRaU9pSTNkbDlOYkdGYVp6RnpiV1pJVGpSYWFtSkVWbXBSUldSelozUkNOV3N5T1ZodVdIUmpTakJaU2tkUkluMC5leUpsZUhRaU9pSjdYQ0p5WlhadlkyRmliR1ZjSWpwY0luUnlkV1ZjSW4waUxDSnpkV0lpT2lKcVptRmpRREF4Wm5rME1UUXdhR3B5WTJvNE1IY3ljRE15Y1hneE9IbHdYQzkxYzJWeWMxd3ZZVzFwY201aFFHcG1jbTluTG1OdmJTSXNJbk5qY0NJNkltRndjR3hwWldRdGNHVnliV2x6YzJsdmJuTmNMM1Z6WlhJaUxDSmhkV1FpT2lJcVFDb2lMQ0pwYzNNaU9pSnFabVpsUURBd01DSXNJbVY0Y0NJNk1UWTBOekkyT0RNd01Dd2lhV0YwSWpveE5qUTNNalUzTlRBd0xDSnFkR2tpT2lKbU1HUTBOR0UxTUMwME1UazBMVFJoWmpRdFltUTFPUzAyWmprek16SXhNalkzWkdZaWZRLktpa1VkVENkRXFMWkdRR2c4TE43LXpZVWNac3dDNmRmU0Jxb2d1RnpvVHhFelFfYmNnZnJTcnVvQWd3MmRERWlPNEJCOEpHel9mUUxWcUVJY1p1RHRvMHc4c0lPREZJdXQyRUVzY3ZxZ2NuRmIyMXFWaTVMRW5FbVFzeW9iSFpPUGJfY081d3JQWEpfUFVsejJQci1iLWN2WjV4UlR0aXBQR1RFM0FuOUdhY19raDBqX2ZLRHRJQXFvQnh4bG1LVERreUJ4MHNwZ0dOTmlfS1VOMzlwaUNnRkg5d1ROcXBZY3o5SW1MX1FacUFxeXQtRnN3M3E1TFJRams4V2tfX0ZEYU9OYmI2enZRS1E4VnctUUR6bmYxeVNuLVRGVFFvaHdoZ0Jab3ZqeUVjYWRLZlR4YXYyUVNsUHNxckhoVFZGQVNnOWRRakg5cDJjV3FqQmRDWG5XdyJ9"
	expectedAccessToken := "eyJ2ZXIiOiIyIiwidHlwIjoiSldUIiwiYWxnIjoiUlMyNTYiLCJraWQiOiI3dl9NbGFaZzFzbWZITjRaamJEVmpRRWRzZ3RCNWsyOVhuWHRjSjBZSkdRIn0.eyJleHQiOiJ7XCJyZXZvY2FibGVcIjpcInRydWVcIn0iLCJzdWIiOiJqZmFjQDAxZnk0MTQwaGpyY2o4MHcycDMycXgxOHlwXC91c2Vyc1wvYW1pcm5hQGpmcm9nLmNvbSIsInNjcCI6ImFwcGxpZWQtcGVybWlzc2lvbnNcL3VzZXIiLCJhdWQiOiIqQCoiLCJpc3MiOiJqZmZlQDAwMCIsImV4cCI6MTY0NzI2ODMwMCwiaWF0IjoxNjQ3MjU3NTAwLCJqdGkiOiJmMGQ0NGE1MC00MTk0LTRhZjQtYmQ1OS02ZjkzMzIxMjY3ZGYifQ.KikUdTCdEqLZGQGg8LN7-zYUcZswC6dfSBqoguFzoTxEzQ_bcgfrSruoAgw2dDEiO4BB8JGz_fQLVqEIcZuDto0w8sIODFIut2EEscvqgcnFb21qVi5LEnEmQsyobHZOPb_cO5wrPXJ_PUlz2Pr-b-cvZ5xRTtipPGTE3An9Gac_kh0j_fKDtIAqoBxxlmKTDkyBx0spgGNNi_KUN39piCgFH9wTNqpYcz9ImL_QZqAqyt-Fsw3q5LRQjk8Wk__FDaONbb6zvQKQ8Vw-QDznf1ySn-TFTQohwhgBZovjyEcadKfTxav2QSlPsqrHhTVFASg9dQjH9p2cWqjBdCXnWw"
	setupCmd := NewEnvSetupCommand().SetBase64Credentials(base64Cred)
	server, err := setupCmd.decodeBase64Credentials()
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8090/", server.Url)
	assert.Equal(t, expectedAccessToken, server.AccessToken)
}
