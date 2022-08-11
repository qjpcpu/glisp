; test different ways of writing an integer
(assert (= 24 0x18))
(assert (= 63 0o77))
(assert (= 13 0b1101))

; test shift operations
(assert (= 4 (sla 1 2)))
(assert (= -1 (sra -4 2)))
(assert (= 2 (sra 4 1)))

; bitwise operations
(assert (= 0b0001 (bit-and 0b0011 0b0101)))
(assert (= 0b0111 (bit-or 0b0011 0b0101)))
(assert (= 0b0110 (bit-xor 0b0011 0b0101)))
(assert (= 0b1100 (bit-and (bit-not 0b0011) 0b1111)))

; arithmetic
(assert (= 5 (+ 3 2)))
(assert (= 2.4 (* 2 1.2)))
(assert (= 2 (mod 5 3)))
(assert (= 1.5 (/ 3 2)))
(assert (= 1.2e3 (* 1.2e2 10)))

(def selection '(1 1.0 0 0.0))

(assert (= '(true true true true) (map number? selection)))
(assert (= '(true false true false) (map int? selection)))
(assert (= '(false true false true) (map float? selection)))
(assert (= '(false false true true) (map zero? selection)))

(assert (= 1.2 (+ 1.1 0.1)))
(assert (= 1.0 (- 1.1 0.1)))
(assert (= 0.11 (* 1.1 0.1)))
(assert (= 1 (- 1.1 0.1)))
(assert (= 5.5 (/ 1.1 0.2)))

(assert (= 1.2 (* 0.2 6)))
(assert (= 1 (float2int (* 0.2 6))))
(assert (= 2 (float2int (* 0.4 6))))

(assert (= -0.1 (- 0 0.1)))

(assert (= 2 (round 1.7)))
(assert (= 2 (round 2.4)))
(assert (= 1 (round 0.5)))

(assert (zero? 0.0000000001))
(assert (not (zero? 0.000000001)))


;; 14.285714285714286
(assert (= "14.286" (float2str 14.285714285714286 3)))
(assert (= "14" (float2str 14.285714285714286 0)))
(assert (= "14" (float2str 14.00000000001 3)))
(assert (= "0" (float2str 0.00000000001 3)))
(assert (= "0" (float2str 0.0 3)))

(assert (= 1.0 1))
(assert (= 97.0 #a))
(assert (= 1 (float2int 1)))
(assert (= "1" (float2str 1)))
(assert (= "1.00" (float2str 1.00)))
(assert (= 1 (round 1)))
(assert (= 2 (/ 10 5)))
(assert (= 18446744073709551614 (- 18446744073709551615 1)))
(assert (= 98 (+ #a 1)))
(assert (= 96 (- #a 1)))
(assert (= 97 (* #a 1)))
(assert (= 97 (/ #a 1)))
(assert (= 97 (* 1.0 #a)))
(assert (= 97 (/ #a 1.0)))
(assert (= 1 (/ 97.0 #a)))
(assert (= 1 (/ #a #a)))
(assert (= 1 (/ 97 #a)))

(assert (= "1.0" (float2str 1.0 -1)))
(assert (= "0" (float2str 0.0 0)))

(assert (= 97 (mod #a #b)))
(assert (= -99 (char2int (bit-not #b))))

(assert (= 2 (sll8 1 1)))
(assert (= 254 (sll8 -1 1)))
(assert (= 2 (sll16 1 1)))
(assert (= 65534 (sll16 -1 1)))
(assert (= 2 (sll32 1 1)))
(assert (= 4294967294 (sll32 -1 1)))
(assert (= 2 (sll64 1 1)))
(assert (= 18446744073709551614 (sll64 -1 1)))

(assert (= 1 (srl8 2 1)))
(assert (= 127 (srl8 -2 1)))
(assert (= 1 (srl16 2 1)))
(assert (= 32767 (srl16 -2 1)))
(assert (= 1 (srl32 2 1)))
(assert (= 2147483647 (srl32 -2 1)))
(assert (= 1 (srl64 2 1)))
(assert (= 9223372036854775807 (srl64 -2 1)))

(assert (= "a" (char2str #a)))

(assert (= #a (int2char 97)))
