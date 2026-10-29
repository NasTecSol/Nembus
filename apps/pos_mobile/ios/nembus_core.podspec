Pod::Spec.new do |s|
  s.name             = 'nembus_core'
  s.version          = '1.0.0'
  s.summary          = 'Nembus Core native framework'
  s.homepage         = 'https://nastecsol.com'
  s.author           = { 'Nembus' => 'admin@nastecsol.com' }
  s.source           = { :path => '.' }
  s.platform         = :ios, '14.0'
  s.vendored_frameworks = 'Frameworks/nembus_core.xcframework'
  s.libraries        = 'resolv', 'pthread'
  s.frameworks       = 'CoreFoundation', 'Security'
end
