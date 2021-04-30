LatestVer="${GITHUB_REF#"refs/tags/"}"
git config --local user.name 'github-actions[bot]'
git config --local user.email '41898282+github-actions[bot]@users.noreply.github.com'
git add --all
git commit -m "chore: Update to $LatestVer"
git tag -d "$LatestVer"
git tag "$LatestVer"
git push
git push --tags -f