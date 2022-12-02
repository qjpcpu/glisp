(defrecord Person
    (Name string)
    (Age int)
    (Hobbits list<string>))

(defrecord Family
    (Father Person)
    (Mother Person)
    (Children list<Person>)
    (Neighbours hash<string,Person>))

(def dad (->Person
          Name "Thomas"
          Age 35))

(def mon (->Person
          Name "Cancy"
          Age 30
          Hobbits '("shopping")))

(def son (->Person
          Name "Link"
          Age 10
          Hobbits '("play game" "reading")))

(def uncle (->Person
            Name "Jackson"
            Age 40))

(def fam (->Family
          Father dad
          Mother mon
          Children (list son)
          Neighbours {"uncle" uncle}))

;; well, start test
;; check father
(assert (= "Thomas"
           (-> fam
               (:Father)
               (:Name))))
(assert (= 35
           (-> (:Father fam)
               (:Age))))
(assert (= nil
           (-> (:Father fam)
               (:Hobbits))))
(assert (= "nothing"
           (-> (:Father fam)
               (:Hobbits "nothing"))))
(assoc dad Hobbits '("driving" "swimming"))
(assert (= '("driving" "swimming")
           (-> (:Father fam)
               (:Hobbits))))

;; check monther
(assert (= "Cancy"
           (-> fam
               (:Mother)
               (:Name))))
(assert (= 30
           (-> (:Mother fam)
               (:Age))))
(assert (= '("shopping")
           (-> (:Mother fam)
               (:Hobbits))))

;; check son
(assoc son Age 11)
(assert (= 11 (:Age son)))

(assoc son Age nil)
(assert (= nil (:Age son)))

;; check neighbours
(assert (= "Jackson"
           (-> (:Neighbours fam)
               (hget "uncle")
               (:Name))))
;; check nothing
(sexp-str fam)

(assert (str/contains? (sexp-str Person) "#class.Person"))
(assert (= "#class.Person" (type Person)))
(assert (= "Person" (type dad)))

(assert (= #`{"Name":"Thomas","Age":35,"Hobbits":["driving","swimming"]}` (json/stringify dad)))

(assert (record? fam))
(assert (record-of? fam Family))

(def multiset (-> (->Person)
    (assoc Name "multi")
    (assoc Age 120)))
(assert (= "multi" (:Name multiset)))
(assert (= 120 (:Age multiset)))


(defrecord WithTag (Int int "age") (Name string "name"))
(def tag-record (->WithTag Int 1 Name "Jack"))
(assert (= "age" (:Int.tag tag-record)))
(assert (= "name" (:Name.tag tag-record)))
(assert (record-class? WithTag))
(assert (= WithTag (get-record-class tag-record)))
(assert (= "WithTag" (hget (record-class-definition (get-record-class tag-record)) "name")))

;; record with default value
(defrecord DefaultRecord (Host string "" "www.example.com") (Port int "" (+ 3305 1)))
(def dr (->DefaultRecord))
(assert (= "www.example.com" (:Host dr)))
(assert (= 3306 (:Port dr)))

(assoc dr (symbol "Host") "github.com")
(assoc dr  "Port" 5432)
(assert (= "github.com" (:Host dr)))
(assert (= 5432 (:Port dr)))
