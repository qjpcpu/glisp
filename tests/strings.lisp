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
(assert (str/end-with? "abc" "bc"))
(assert (str/contains? "abc" "b"))
(assert (= "Abc" (str/title "abc")))
(assert (= "abc" (str/lower "ABC")))
(assert (= "ABC" (str/upper "abc")))
(assert (= "aBBc" (str/replace "abc" "b" "BB")))
(assert (= "bc" (str/trim-prefix "abc" "a")))
(assert (= "ab" (str/trim-suffix "abc" "c")))
(assert (= "abc" (str/trim-space " abc ")))
(assert (= 2 (str/count " abc cd" "c")))
(assert (= 1 (str/count "abc cd" "b")))

(assert (= 1024 (str2int "1024")))

(assert (= "1024" (str 1024)))
(assert (= "true" (str true)))

(assert (= ["a" "b"] (str/split "a b" " ")))

(assert (= "a_b" (str/join ["a" "b"] "_")))


(assert (str/digit? "0234"))
(assert (not (str/digit? "j0234")))

(assert (str/alpha? "abC"))
(assert (not (str/alpha? "1abC")))

(assert (str/title? "Abc"))
(assert (str/title? "A"))
(assert (not (str/title? "aBc")))
(assert (not (str/title? "")))

(assert (= 1.1 (str2float "1.1")))
(assert (= 1 (str2float "1")))

(assert (= "51225783c75fde283cf746a2904c7920" (str/md5 "glisp")))

(assert (= "语言*" (str/mask "语言学" 2 1 "*")))
(assert (= "语言*" (str/mask "语言学" 2 100 "*")))
(assert (= "语言*" (str/mask "语言学" 2 -1 "*")))
(assert (= "语言学" (str/mask "语言学" 20 1 "*")))
(assert (= "语l**学" (str/mask "语lan学" 2 2 "*")))
