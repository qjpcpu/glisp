(defn list/complement [a b]
  (let [h (foldl (fn [e acc] (hset! acc e '()) acc) {} b)]
       (filter (fn [e] (not (exist? h e))) a)))

(defn list/intersect [a b]
  (let [h (foldl (fn [e acc] (hset! acc e '()) acc) {} b)]
       (filter (fn [e] (exist? h e)) a)))

(defn list/uniq [a]
  (let [h {}]
       (filter (fn [e]
                   (cond (exist? h e) false
                         (begin (hset! h e '()) true))) a)))

(defn list/union [a b] (concat a b))
