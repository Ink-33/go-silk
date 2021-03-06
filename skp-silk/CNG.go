package skpsilk

import "math"

//silk/src/SKP_Silk_CNG.c

func CNG_exc(residual []int16, exc_buf_Q10 []int32, Gain_Q16 int32, length int, rand_seed *int32) {
	var seed int32
	var i, idx, exc_mask int
	exc_mask = CNG_BUF_MASK_MAX
	for exc_mask > length {
		exc_mask = RSHIFT(exc_mask, 1)
	}
	seed = *rand_seed
	for i = 0; i < length; i++ {
		seed = RAND(seed)
		idx = (int)(RSHIFT(int(seed), 24) & exc_mask)
		assert(idx >= 0)
		assert(idx <= CNG_BUF_MASK_MAX)
		residual[i] = int16(SAT16(int16(RSHIFT_ROUND(SMULWW(exc_buf_Q10[idx], Gain_Q16), 10))))
	}
	*rand_seed = seed
}

func CNG_Reset(psDec *decoder_state) {
	var i, NLSF_step_Q15, NLSF_acc_Q15 int
	NLSF_step_Q15 = (math.MaxInt16) / (psDec.LPC_order + 1)
	NLSF_acc_Q15 = 0
	for i = 0; i < psDec.LPC_order; i++ {
		NLSF_acc_Q15 += NLSF_step_Q15
		psDec.sCNG.CNG_smth_NLSF_Q15[i] = NLSF_acc_Q15
	}
	psDec.sCNG.CNG_smth_Gain_Q16 = 0
	psDec.sCNG.rand_seed = 3176576
}

func CNG(psDec *decoder_state, psDecCtrl *decoder_control, signal []int16, length int) {
	var i, subfr int
	var tmp_32, Gain_Q26, max_Gain_Q16 int32
	var LPC_buf = make([]int16, 16)
	var CNG_sig = make([]int16, 16)
	var psCNG = &CNG_struct{}
	if psDec.fs_kHz != psCNG.fs_kHz {
		/* Reset state */
		CNG_Reset(psDec)

		psCNG.fs_kHz = psDec.fs_kHz
	}
	if psDec.ossCnt == 0 && psDec.vadFlag == NO_VOICE_ACTIVITY {
		/* Update CNG parameters */

		/* Smoothing of LSF's  */
		for i = 0; i < psDec.LPC_order; i++ {
			psCNG.CNG_smth_NLSF_Q15[i] += SMULWB(psDec.prevNLSF_Q15[i]-psCNG.CNG_smth_NLSF_Q15[i], CNG_NLSF_SMTH_Q16)
		}
		/* Find the subframe with the highest gain */
		max_Gain_Q16 = 0
		subfr = 0
		for i = 0; i < NB_SUBFR; i++ {
			if psDecCtrl.Gains_Q16[i] > max_Gain_Q16 {
				max_Gain_Q16 = psDecCtrl.Gains_Q16[i]
				subfr = i
			}
		}
		//TODO
		/* Update CNG excitation buffer with excitation from this subframe */
		SKP_memmove(&psCNG.CNG_exc_buf_Q10[psDec.subfr_length], psCNG.CNG_exc_buf_Q10, (NB_SUBFR-1)*psDec.subfr_length*sizeof(SKP_int32))
		SKP_memcpy(psCNG.CNG_exc_buf_Q10, &psDec.exc_Q10[subfr*psDec.subfr_length], psDec.subfr_length*sizeof(SKP_int32))

		/* Smooth gains */
		for i = 0; i < NB_SUBFR; i++ {
			psCNG.CNG_smth_Gain_Q16 += SMULWB(psDecCtrl.Gains_Q16[i]-psCNG.CNG_smth_Gain_Q16, CNG_GAIN_SMTH_Q16)
		}
	}
	/* Add CNG when packet is lost and / or when low speech activity */
	if psDec.lossCnt { //|| psDec->vadFlag == NO_VOICE_ACTIVITY ) {

		/* Generate CNG excitation */
		CNG_exc(CNG_sig, psCNG.CNG_exc_buf_Q10, psCNG.CNG_smth_Gain_Q16, length, &psCNG.rand_seed)

		/* Convert CNG NLSF to filter representation */
		NLSF2A_stable(LPC_buf, psCNG.CNG_smth_NLSF_Q15, psDec.LPC_order)

		Gain_Q26 = int32(1 << 26) /* 1.0 */

		/* Generate CNG signal, by synthesis filtering */
		if psDec.LPC_order == 16 {
			LPC_synthesis_order16(CNG_sig, LPC_buf, Gain_Q26, psCNG.CNG_synth_state, CNG_sig, length)
		} else {
			LPC_synthesis_filter(CNG_sig, LPC_buf, Gain_Q26, psCNG.CNG_synth_state, CNG_sig, length, psDec.LPC_order)
		}
		/* Mix with signal */
		for i = 0; i < length; i++ {
			tmp_32 = int32(signal[i] + CNG_sig[i])
			signal[i] = SAT16(int16(tmp_32))
		}
	} else {
		//TODO
		SKP_memset(psCNG.CNG_synth_state, 0, psDec.LPC_order*sizeof(SKP_int32))
	}
}
