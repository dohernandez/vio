# Init script to kick-start your project
url=$(git remote get-url origin)

url_nopro=${url#*//}
url_noatsign=${url_nopro#*@}

gh_repo=${url_noatsign#"github.com:"}
gh_repo=${gh_repo#"github.com/"}
gh_repo=${gh_repo%".git"}

copyright="$(date +%Y) $(git config user.name)"
project_name=$(basename $gh_repo)
project_name_camel_case=$(echo "$project_name" | perl -pe 's/(^|_|-| )./uc($&)/ge;s/_|-| //g')
project_name_lower_camel_case=$(echo "$project_name_camel_case" |  perl -nE 'say lcfirst')
project_name_trim=$(echo "$project_name_lower_camel_case" | perl -pe 's/./lc($&)/ge')
project_name_title=$(echo "$project_name" | perl -pe 's/(^|_|-| )./uc($&)/ge;s/_|-/ /g')
project_name_snake_case=$(echo "$project_name" | perl -pe 's/./lc($&)/ge;s/-| /_/g')


echo "## Replacing all kit-template references by $project_name"
find . -type f -not -name run_me.sh -not -path "./.git/*" -print0 | xargs -0 perl -i -pe "s|2021 dohernandez|$copyright|g"
find . -type f -not -name run_me.sh -not -path "./.git/*" -print0 | xargs -0 perl -i -pe "s|dohernandez/kit-template|$gh_repo|g"
find . -type f -not -name run_me.sh -not -path "./.git/*" -print0 | xargs -0 perl -i -pe "s|kit-template|$project_name|g"
echo "## Replacing all KitTemplate references by $project_name_camel_case"
find . -type f -not -name run_me.sh -not -path "./.git/*" -print0 | xargs -0 perl -i -pe "s|KitTemplate|$project_name_camel_case|g"
echo "## Replacing all kitTemplate references by $project_name_lower_camel_case"
find . -type f -not -name run_me.sh -not -path "./.git/*" -print0 | xargs -0 perl -i -pe "s|kitTemplate|$project_name_lower_camel_case|g"
echo "## Replacing all kittemplate references by $project_name_trim"
find . -type f -not -name run_me.sh -not -path "./.git/*" -print0 | xargs -0 perl -i -pe "s|kittemplate|$project_name_trim|g"
echo "## Replacing all Kit Template references by $project_name_title"
find . -type f -not -name run_me.sh -not -path "./.git/*" -print0 | xargs -0 perl -i -pe "s|Kit Template|$project_name_title|g"
echo "## Replacing all kit_template references by $project_name_snake_case"
find . -type f -not -name run_me.sh -not -path "./.git/*" -print0 | xargs -0 perl -i -pe "s|kit_template|$project_name_snake_case|g"

echo "## Renaming cmd/kit-template to cmd/$project_name"
mv cmd/kit-template "cmd/$project_name"
git add cmd/

echo "## Updating README.md from README.md.template"
rm ./README.md
mv ./README.md.template ./README.md

echo "## Removing this script"
rm ./run_me.sh

echo "## Please check the @TODO's:"
git grep TODO | grep -v run_me.sh | grep -v "resources/proto/protoc-gen-openapiv2/*" | grep -v "resources/proto/google/*"