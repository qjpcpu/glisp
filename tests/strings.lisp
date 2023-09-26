(assert (= "abc" (append "ab" #c)))
(assert (= "abcd" (concat "ab" "cd")))
(assert (= "abcdef" (concat "ab" "cd" "ef")))
(assert (= "bc" (slice "abcd" 1 3)))
(assert (= "bcd" (slice "abcd" 1)))
(assert (= #c (sget "abcd" 2)))
(assert (= 3 (len "abc")))

(assert (string? "asdfsdaf"))
(assert (char? #c))
(assert (symbol? 'a))

(assert (str/start-with? "abc" "ab"))
(assert (not (str/start-with? "abc" "aB")))
(assert (str/start-with? "abc" "aB" true))
(assert (str/end-with? "abc" "bc"))
(assert (str/contains? "abc" "b"))
(assert (str/contains? "abc" "B" true))
(assert (= "Abc" (str/title "abc")))
(assert (= "abc" (str/lower "ABC")))
(assert (= "ABC" (str/upper "abc")))
(assert (= "aBBc" (str/replace "abc" "b" "BB")))
(assert (= "aBbc" (str/replace "abbc" "b" "B" 1)))
(assert (= "bc" (str/trim-prefix "abc" "a")))
(assert (= "ab" (str/trim-suffix "abc" "c")))
(assert (= "abc" (str/trim-space " abc ")))
(assert (= 2 (str/count " abc cd" "c")))
(assert (= 1 (str/count "abc cd" "b")))

(assert (= 1024 (int "1024")))

(assert (= "1024" (string 1024)))
(assert (= "true" (string true)))

(assert (str/equal-fold? "abc" "AbC"))
(assert (= ["a" "b"] (str/split "a b" " ")))
(assert (= ["a" "b c d"] (str/split "a b c d" " " 2)))

(assert (= "a_b" (str/join ["a" "b"] "_")))
(assert (= "a_b" (str/join '("a" "b") "_")))
(assert (= "" (str/join '() "_")))
(assert (= "a_b" (str/join2 "_" '("a" "b"))))


(assert (str/digit? "0234"))
(assert (not (str/digit? "j0234")))

(assert (str/alpha? "abC"))
(assert (not (str/alpha? "1abC")))

(assert (str/title? "Abc"))
(assert (str/title? "A"))
(assert (not (str/title? "aBc")))
(assert (not (str/title? "")))

(assert (= 1.1 (float "1.1")))
(assert (= 1 (float "1")))

(assert (= "51225783c75fde283cf746a2904c7920" (str/md5 "glisp")))

(assert (= "语言*" (str/mask "语言学" 2 1 "*")))
(assert (= "语言*" (str/mask "语言学" 2 100 "*")))
(assert (= "语言*" (str/mask "语言学" 2 -1 "*")))
(assert (= "语言学" (str/mask "语言学" 20 1 "*")))
(assert (= "语l**学" (str/mask "语lan学" 2 2 "*")))

(assert (str/integer? "123"))
(assert (str/float? "0.123"))
(assert (str/bool? "false"))

;; raw string
(assert (= "start\\n\\t\\e# \" \" '' ~ ~@ " #`start\n\t\e# " " '' ~ ~@ `))
(assert (= "line1\n           line2\n           line3" #`line1
           line2
           line3`))

(assert (empty? ""))
(assert (empty? (bytes "")))

(assert (= 3 (len "中国人")))

(assert (bool "true"))
(assert (not (bool "false")))
(assert (bool true))
(assert (not (bool false)))

(assert (= #a (char "a")))

(assert (= "abc" (string "abc")))

;; unicode string

(assert (= 7 (len "I'am中国人")))
(assert (= 13 (len (bytes "I'am中国人"))))
(assert (= "m中" (slice "I'am中国人" 3 5)))
(assert (= "a美" (append "a" (char "美"))))
(assert (= 4 (str/index "I'am中国人" "中")))
(assert (= 5 (str/index "I'am中国人" "国")))
(assert (= 0 (str/index "I'am中国人" "I")))
(assert (= 0 (str/index "美I'am中国人" "美")))
(assert (= -1 (str/index "I'am中国人" "x")))

(assert (= "" (string nil)))
(assert (= "1" (string 1)))
(assert (= "1.2" (string 1.2)))
(assert (= "F" (string #F)))
(assert (= "" (string [])))
(assert (= "s" (string '("s"))))
(assert (= "abc" (string '(#a #b #c))))
(assert (= "abc" (string '[#a #b #c])))
(assert (= "atrueb1233.14BYTESd" (string ['("a" true "b") [1 2 3 3.14 (bytes "BYTES")] #d])))

(assert (= "AAA" (str/repeat "A" 3)))
