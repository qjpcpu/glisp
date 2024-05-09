(defn str/first [& strs]
  "Usage: (str/first & strs)
Return first non empty str or nil."
  (or  (->> (stream strs)
            (reject #(or (not (string? %)) (empty? %)))
            (take 1)
            (realize)
            (car)) ""))
