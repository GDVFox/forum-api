
SELECT p.id, p.message, p.is_edited, p.created, p.author, p.forum, p.thread, p.parent
FROM posts p WHERE thread = 353 AND p.id > 0 ORDER BY p.created LIMIT 65;

SELECT p.id, p.message, p.is_edited, p.created, p.author,
p.forum, p.thread, p.parent FROM posts p WHERE p.thread = 2118
                                           AND (p.created, p.id) > (SELECT posts.created, posts.id FROM posts WHERE posts.id=20713) ORDER BY (p.created, p.id);

SELECT p.id, p.message, p.is_edited, p.created, p.author, p.forum, p.thread, p.parent FROM posts p WHERE parent IN (
  	SELECT p.id
		FROM posts p WHERE thread = 353 AND p.id > 0 AND p.parent IS NULL
);


WITH parents AS (
	SELECT * FROM posts WHERE thread = 353 AND posts.id > 0 AND posts.parent IS NULL ORDER BY posts.created
)
SELECT p.id, p.message, p.is_edited, p.created, p.author, p.forum, p.thread, p.parent FROM posts p INNER JOIN parents ON p.parent = parents.id ORDER BY p.parent;

SELECT p.id, p.message, p.is_edited, p.created, p.author, p.forum, p.thread, p.parent FROM posts p JOIN (
	SELECT posts.id FROM posts WHERE thread = 353 AND posts.id > 0 AND posts.parent IS NULL ORDER BY posts.created
) parents ON p.parent = parents.id OR p.id = parents.id ORDER BY p.parent;

WITH RECURSIVE recursetree (id, message, is_edited, created, author, forum, thread, parent, path, depth) AS (
    SELECT *, array[]::bigint[], 0 FROM posts
    WHERE posts.parent IS NULL AND thread = 118 AND posts.id > 0
  UNION ALL
    SELECT p.id, p.message, p.is_edited, p.created, p.author, p.forum, p.thread, p.parent, array_append(path, p.parent), depth+1
    FROM posts p
    JOIN recursetree rt ON rt.id = p.parent
  )
SELECT rt.id, rt.message, rt.is_edited, rt.created, rt.author, rt.forum, rt.thread,
       rt.parent, rt.path, rt.depth FROM recursetree rt ORDER BY (rt.path, rt.id) LIMIT 30;

WITH RECURSIVE recursetree (id, message, is_edited, created, author, forum, thread, parent, path) AS (
	(SELECT *, array[id] FROM posts
    WHERE posts.parent IS NULL AND thread = 599 AND posts.id > 0 LIMIT 1)
  UNION ALL
    SELECT p.id, p.message, p.is_edited, p.created, p.author, p.forum, p.thread, p.parent, array_append(path, p.parent)
    FROM posts p
    JOIN recursetree rt ON rt.id = p.parent
  )
SELECT rt.id, rt.message, rt.is_edited, rt.created, rt.author, rt.forum, rt.thread,
       rt.parent FROM recursetree rt ORDER BY rt.path[1], rt.path, rt.id;


SELECT p.id, p.message, p.is_edited, p.created, p.author,
       p.forum, p.thread, p.parent, parents FROM posts p
WHERE p.thread = 1116 AND p.parents > (SELECT posts.parents FROM posts WHERE posts.id = 6036) ORDER BY p.parents LIMIT 3;

SELECT p.id, p.message, p.is_edited, p.created, p.author,
       p.forum, p.thread, p.parent, parents FROM posts p WHERE p.parents[1] IN (
         SELECT posts.id FROM posts WHERE posts.thread = 380 AND posts.parent IS NULL
																		LIMIT 3
      )  ORDER BY p.parents;

SELECT p.id, p.message, p.is_edited, p.created, p.author,
       p.forum, p.thread, p.parent, parents FROM posts p WHERE p.parents[1] IN (
         SELECT posts.id FROM posts WHERE posts.thread = 127 AND posts.parent IS NULL
                                      AND posts.id < (SELECT COALESCE(posts.parent, posts.id) FROM posts WHERE posts.id = 2193) ORDER BY posts.id DESC LIMIT 3
      ) ORDER BY p.parents[1] DESC, p.parents;

SELECT posts.id FROM posts WHERE posts.thread = 127 AND posts.parent IS NULL
                                      AND posts.parent < (SELECT posts.parent FROM posts WHERE posts.id = 2201)