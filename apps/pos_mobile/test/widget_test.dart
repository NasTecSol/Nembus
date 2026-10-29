import 'package:flutter_test/flutter_test.dart';
import 'package:pos_mobile/main.dart';

void main() {
  testWidgets('Nembus Mobile POS splash screen renders', (WidgetTester tester) async {
    await tester.pumpWidget(const NembusMobileApp());
    expect(find.text('NEMBUS POS'), findsOneWidget);
    await tester.pump(const Duration(milliseconds: 100));
    await tester.pump(const Duration(seconds: 3));
  });
}
