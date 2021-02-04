# Simple ERD Example

```erd {cmd=true stdin=true args=["-f", "svg"] output="html"}

colors {
    person: "#fcecec",
    loc: "#ececfc",
}

[User.Person] {bgcolor: "person" label: "comment"}
*name
height
weight
+birth_location_id

[Location] {bgcolor: "loc"}
*id
city
state
country

User.Person 1--* Location
```
