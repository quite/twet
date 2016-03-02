
# A lot to do

* cli: posting possible, a hook for syncing is still missing
* http: fetching has been parallelized but may need some tidying
* cli/output: can limit timeline tweets by duration; more?
* cli/output: ascending/descending time order
* http: think about redirect, and handling of 401, 301, 404?
* cli/http: a "follow" command should probably resolve 301s (cache-control or not?)
* cli: add/remove who you're following
* cache: behaviour when adding/removing following
* following: require unique URL?
* http/cli: setting of user-agent to be optional?
* Currently, one is supposed to follow oneself -- is that good design or not
  (once we start dealing with twtfile)
* Thought: Hm, I shall not require neither nick nor twturl.
* ...
