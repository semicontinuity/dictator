package main

import (
	kbd "github.com/micmonay/keybd_event"
	log "github.com/sirupsen/logrus"
	"time"
)

var keyMapping = map[rune]int{
	' ': kbd.VK_SPACE,
	'a': kbd.VK_A,
	'b': kbd.VK_B,
	'c': kbd.VK_C,
	'd': kbd.VK_D,
	'e': kbd.VK_E,
	'f': kbd.VK_F,
	'g': kbd.VK_G,
	'h': kbd.VK_H,
	'i': kbd.VK_I,
	'j': kbd.VK_J,
	'k': kbd.VK_K,
	'l': kbd.VK_L,
	'm': kbd.VK_M,
	'n': kbd.VK_N,
	'o': kbd.VK_O,
	'p': kbd.VK_P,
	'q': kbd.VK_Q,
	'r': kbd.VK_R,
	's': kbd.VK_S,
	't': kbd.VK_T,
	'u': kbd.VK_U,
	'v': kbd.VK_V,
	'w': kbd.VK_W,
	'x': kbd.VK_X,
	'y': kbd.VK_Y,
	'z': kbd.VK_Z,
}

func typeKeys(kb kbd.KeyBonding, textStream chan string) {
	for text := range textStream {
		log.Infof("Typing text: %s", text)
		time.Sleep(500 * time.Millisecond)

		//kb.Clear()

		for _, char := range text {
			//log.Infof("Char: %d", char)

			vk, ok := keyMapping[char]
			if ok {
				//log.Infof("VK: %d", vk)

				//kb.AddKey(vk)


				kb.SetKeys(vk)

				err := kb.Press()
				if err != nil {
					panic(any(err))
				}

//				time.Sleep(10 * time.Millisecond)

				err = kb.Release()
				if err != nil {
					panic(any(err))
				}

//				time.Sleep(10 * time.Millisecond)

				kb.Clear()
			}
		}

		//kb.Launching()

/*		kb.Press()
		time.Sleep(10 * time.Millisecond)
		kb.Release()
		time.Sleep(10 * time.Millisecond)

		kb.Clear()
*/
		//kb.Launching()
	}
}
