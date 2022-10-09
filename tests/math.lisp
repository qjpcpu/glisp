(assert (= 127 (srl8 -1 1)))
(assert (= 7 (bit-or 1 2 4)))
(assert (= 2 (bit-and 7 2 6)))

(assert (= "0x41" (0x 65)))
(assert (= "0o755" (0o 493)))
(assert (= "0b1110" (0b 14)))
