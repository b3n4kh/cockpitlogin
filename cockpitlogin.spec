Name:           cockpitlogin
Version:        0
Release:        1
Summary:        Cockpit-project TLS Client Cert Login Broker
License:        Beerware
URL:            https://github.com/b3n4kh/cockpitlogin
Source0:        https://github.com/b3n4kh/cockpitlogin/releases/download/%{Version}.%{Release}/cockpitlogin.tar.gz
Source1:        systemd/%{Name}.service
Source2:        systemd/%{Name}.socket


# If go_compiler is not set to 1, there is no virtual provide. Use golang instead.
BuildRequires:  %{?go_compiler:compiler(go-compiler)}%{!?go_compiler:golang}

%prep
%setup -q -n %{repo}-%{commit}

%install
install -d %{buildroot}%{_bindir}
install -p -m 755 %{Name} %{buildroot}%{_bindir}

%{_bindir}/%{Name}

