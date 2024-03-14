//TODO 
//KAN vi lage en config fil?
    Philip: Config fil til hva?
//Kan vi sette ned tiden det tar å sjekke at vi er offline? 
//jeg gjode slik at vi sender mld ekstremt mye oftere. gjør systemet litt raskere
    Philip: Sikkert fint det, så lenge vi ikke DOS-er nettet
//vi trenger ikke func IsStateCorrupted, men skader det å ha den? 
    Philip: Det skader ikke, men er det overflødig, så burde det ikke være med. Er det en ekstra sjekk for noe som aldri vil oppstå, så er det jo ikke noe vits.
//dersom vi fjerner den kan vi også fjerne isValidBehavior() og isValidDirection()
    Philip: Ja var det jeg tenkte også. Tror det er nice



//komentarer vi har fått fra peer reviws. 
//Mange filer, kan bli uoversiktelig
    Philip: Det er selvfølgelig personlig preferanse, men flere filer betyr gjerne mindre scrolling. Og det er som regel (ikke alltid) god praktisk å splitte funksjonalitet i forskjellige filer.
//mainen er ikke så lett å skjønne seg på. - gjøre main til egen folder. Kan init nodes bli omdøpt til main? 
    Philip: main er et reservert funksjonsnavn, så burde ikke gjøres. Kan ev. endres til mainLogic slik det var, men det er like mye, om ikke mindre beskrivende.

//simen du har bedre peiling enn meg, men fsm står jo for FinalStateMachine, hvorfor heter mappen FSM når det ikke er en FSM. Burde den kanskje omdøpes til noe mer passende som 
    Philip: Det står for Finite State Machine, som bare betyr at State Machinen kan ha én mulig state av gangen, fra et gitt, endelig sett av mulige states.

All the source files for your code and a README containing simple instructions for how to compile/execute your code should be handed in as one .zip file named  "code-##.zip", where ## is your group ID. Do not include executables or any names or other identifications in your code.
    Philip: Jeg fikser nå (Døper denne filen her til spm.md siden README.md er reservert i GitHub)

The code you hand in will be evaluated qualitatively in the code review and functionally during the Factory Acceptance Test (FAT). In other words, it is this code that you have to run during the FAT, so make sure the code you hand in works.