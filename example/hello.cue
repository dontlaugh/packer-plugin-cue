package hello



world: string | "yes!" @tag(world,type=string)

expect_bytes: '''
#!/bin/bash

echo "amazing"

'''

expect_json: {
    i_am: "json"
    one: {
        "two": ["three", "three", "three"]
    }
}


expect_yaml: {
    i_am: "yaml"
    one: {
        "two": ["three", "three", "three"]
    }
}


expect_toml: {
    i_am: "toml"
    one: {
        "two": ["three", "three", "three"]
    }
}
