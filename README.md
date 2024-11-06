<img src=".assets/lay.jpg" alt="Electronics photo" style="width: 400px;" align="right">

# lay 📥
📥 Lay helps with installing or compiling applications 📥

There are many ways to install and compile software. It becomes very hard to port from one environment to another the
processes that are required to install that piece of software. Lay is a tool that abstracts the myriad of ways an
application or a package can be added to a system.

So instead of looking for the right package name or the right command to pass to install a software, lay takes care of it.

## 🙌 Benefits
- No more need to look up the different names for packages
- Takes care of using the most efficient way to compile the software
- Using [hery](https://github.com/AmadlaOrg/hery) the requirements for the compilation and of a package can be defined making it portable

Lay supports Windows and Linux.

**Mac OS X:**
> Amadla focuses on server environments mostly. So for this reason Mac OS X is not supported. That said, nothing is
> stopping anyone from contributing support for Mac OS X if need be.

**Other:**
> There are many other OSs out there that could be supported. But since there is a limited amount of time they haven't
> been added yet. But nothing is set in stone. So anyone can add their contributions.

## Linux
### Package managers supported
- [`apt`](https://wiki.debian.org/Apt)
- [`dpkg`](https://wiki.debian.org/Teams/Dpkg)
- [`yum`](http://yum.baseurl.org/)
- [`dnf`](https://docs.fedoraproject.org/en-US/quick-docs/dnf/)
- [`rpm`](https://rpm.org/)

### Containers
- [Docker](https://www.docker.com/)
- [Podman](https://podman.io/)

### Compiling
- [Autotools](https://www.gnu.org/software/automake/manual/html_node/Autotools-Introduction.html)

## Windows
### Package managers supported
- [Chocolatey](https://chocolatey.org/)
- [winget](https://winget.run/)
- [scoop](https://scoop.sh/)

## How it works
Every thing is stored in different [hery](https://github.com/AmadlaOrg/hery) entity configuration file that define the
requirements of a software for each different ways it is installed. They can all be overwritten. It also defines what
are the equivalent in the different package managers.

> Since every piece of software needs to have their own entity configuration it is impossible to support all of them.
> But as many as possible especially the most popular for server environments.

## 🕹️ Usage
There are different ways to use this tool. The main way to use it is with Amadla/hery but to keep things flexible it is
also possible to specify what software and how it needs to be installed.

> Amadla [Entity Application](https://github.com/AmadlaOrg/EntityApplication) is the entity that is supported.

### Help
```bash
lay --help|-h
```

### With Entity (Amadla)
Amadla [Entity Application](https://github.com/AmadlaOrg/EntityApplication) is the entity that is supported.

```bash
lay entity <collection name> <entity name>
```

Example:
```bash
lay entity amadla MyApplications
```

### Package manager
```bash
lay package <package manager name> <collection name> <entity name> <software name>
```

### Compile
```bash
lay compile <collection name> <entity name> <software name>
```

```bash
lay compile <entity path> <software name>
```

### Container
```bash
lay container --engine|-e <container engine> <collection name> <entity name> <software name>
```

```bash
lay container --engine|-e <container engine> <entity path> <software name>
```

## Env variables
```bash
LAY_CONTAINER_ENGINE="podman"
```

## ©️ Copyright
- "<a rel="noopener noreferrer" href="https://www.flickr.com/photos/37667416@N04/3987005186">&#039;&#039;All the light and life of day came on ; and amidst it all, and pressing down the grass whose every blade bore twenty tiny lives, lay the dead man, with his stark and rigid face turned upward to the sky.&#039;&#039;</a>" by <a rel="noopener noreferrer" href="https://www.flickr.com/photos/37667416@N04">Biblioteca Rector Machado y Nuñez</a> is marked with <a rel="noopener noreferrer" href="https://creativecommons.org/publicdomain/mark/1.0/?ref=openverse">Public Domain Mark 1.0 <img src="https://mirrors.creativecommons.org/presskit/icons/pd.svg" style="height: 1em; margin-right: 0.125em; display: inline;" /></a>.

## :scroll: License

The license for the code and documentation can be found in the [LICENSE](./LICENSE) file.

---

Made in Québec 🏴󠁣󠁡󠁱󠁣󠁿, Canada 🇨🇦!
