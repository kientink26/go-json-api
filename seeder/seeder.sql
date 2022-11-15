INSERT INTO movies (title, year, runtime, genres)
VALUES
    ('One Flew Over the Cuckoo''s Nest',1975,133,'{drama}'),
    ('Avengers: Infinity War',2018,150,'{action,adventure,sci-fi}'),
    ('Me Before You',2016,110,'{drama,romance}'),
    ('Casablanca',1942,102,'{drama,romance,war}'),
    ('Moana',2016,107,'{animation,adventure}'),
    ('Black Panther',2018,134,'{action,adventure}'),
    ('Deadpool',2016,108,'{action,comedy}'),
    ('The Breakfast Club',1986,96,'{drama}');

-- username: admin@example.com
-- password: pa55word
INSERT INTO users (name, email, password_hash, activated)
VALUES
    ('Kiên Phạm','admin@example.com','$2a$12$NK/xdRYy5OfT8o6YEiTvI.uqXU602uC3ns36Fc1Eoskx8kWoT5KUC',true);

INSERT INTO users_permissions
SELECT (SELECT users.id FROM users WHERE users.email = 'admin@example.com')
     ,permissions.id FROM permissions WHERE permissions.code = ANY('{users:read,permissions:read,permissions:write}');

