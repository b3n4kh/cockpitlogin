%define version 1
%define release 2
%define name cockpitlogin
%define debug_package %{nil}
%define _build_id_links none

Name:           %{name}
Version:        %{version}
Release:        %{release}
Summary:        Cockpit-project TLS Client Cert Login Broker
License:        Beerware
URL:            https://github.com/b3n4kh/cockpitlogin
Source0:        %{name}-%{version}.%{release}.tar.gz

ExclusiveArch:  %{go_arches}
Requires: systemd nginx
BuildRequires: systemd
Requires(pre): shadow-utils

%description
Cockpit-project TLS Client Cert Login Broker

%prep
%setup -n %{name}

%pre
/usr/bin/getent passwd %{name} > /dev/null 2>&1 || /usr/sbin/useradd -r -M -u 1896 -s /sbin/nologin %{name}
/usr/bin/getent group nginx > /dev/null 2>&1 && /usr/sbin/usermod -aG nginx %{name}

%post
%systemd_post %{name}.service
/usr/sbin/semanage fcontext -a -t httpd_var_run_t "/var/run/cockpitlogin(/.*)?" 2>/dev/null || /usr/bin/true
/usr/sbin/restorecon -R /var/run/cockpitlogin || /usr/bin/true

%build
mkdir -p ./_build/src/github.com/b3n4kh/
ln -s $(pwd) ./_build/src/github.com/b3n4kh/%{name}

export GOPATH=$(pwd)/_build:%{gopath}
go build -o bin/%{name} .

%install
install -d %{buildroot}%{_bindir}
install -d %{buildroot}%{_sysconfdir}/%{name}
install -d %{buildroot}%{_unitdir}
install -p -m 755 bin/%{name} %{buildroot}%{_bindir}
install -p -m 644 systemd/%{name}.service %{buildroot}%{_unitdir}
install -p -m 644 config.json %{buildroot}%{_sysconfdir}/%{name}

%files
%{_bindir}/%{name}
%{_unitdir}/%{name}.service
%config(noreplace) %{_sysconfdir}/%{name}/config.json

