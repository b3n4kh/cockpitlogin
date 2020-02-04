%define version 1
%define release 0
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
BuildRequires: systemd

%description
Cockpit-project TLS Client Cert Login Broker

%prep
%setup -n %{name}

%build
mkdir -p ./_build/src/github.com/b3n4kh/
ln -s $(pwd) ./_build/src/github.com/b3n4kh/%{name}

export GOPATH=$(pwd)/_build:%{gopath}
go build -o bin/%{name} .

%install
install -d %{buildroot}%{_bindir}
install -d %{buildroot}%{_unitdir}
install -p -m 755 bin/%{name} %{buildroot}%{_bindir}
install -p -m 644 systemd/%{name}.service %{buildroot}%{_unitdir}
install -p -m 644 systemd/%{name}.socket %{buildroot}%{_unitdir}



%files
%{_bindir}/%{name}
%{_unitdir}/%{name}.service
%{_unitdir}/%{name}.socket
