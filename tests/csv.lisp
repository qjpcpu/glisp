;; file
(def file "not-exist-file.csv")
(os/remove-file file)

(def content [["a" "b" "c"]])
(csv/write-file file content)
(assert (= content (csv/read-file file)))
(assert (= "a,b,c\n" (string (os/read-file file))))

(def content [["a" "b" "c"] ["1" "2" "3"]])
(csv/write-file file content)
(assert (= content (csv/read-file file)))
(assert (= "a,b,c\n1,2,3\n" (string (os/read-file file))))

(def content [{"a" "1" "b" "2" "c" "3"} {"a" "4" "b" "5" "c" "6"}])
(csv/write-file file content)
(assert (= (json/stringify  content) (json/stringify (csv/read-file file 'hash))))
(assert (= "a,b,c\n1,2,3\n4,5,6\n" (string (os/read-file file))))

(def content [{"a" [1 2 3] "b" {"k" 1} "c" 3}])
(csv/write-file file content)
(assert (= #`[["a","b","c"],["[1,2,3]","{\"k\":1}","3"]]` (json/stringify (csv/read-file file))))

(os/remove-file file)
